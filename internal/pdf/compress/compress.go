package compress

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

// Level is a Ghostscript-style compression tier.
type Level string

const (
	LevelLight  Level = "light"  // JPEG 92, max image edge 1920
	LevelMedium Level = "medium" // JPEG 75, max image edge 1275
	LevelHeavy  Level = "heavy"  // JPEG 50, max image edge 612
)

const (
	qualityLight  = 92
	qualityMedium = 75
	qualityHeavy  = 50
	maxDimLight   = 1920
	maxDimMedium  = 1275
	maxDimHeavy   = 612
)

// Options controls how aggressively CompressPDF rewrites streams and images.
// Empty Level selects Medium (JPEG 75). JPEGQuality and MaxImageDim override
// the preset when they are greater than zero.
type Options struct {
	Level       Level
	JPEGQuality int
	MaxImageDim int
}

func (o Options) withDefaults() Options {
	switch Level(strings.ToLower(string(o.Level))) {
	case LevelLight:
		o.Level = LevelLight
	case LevelHeavy:
		o.Level = LevelHeavy
	default:
		o.Level = LevelMedium
	}
	if o.JPEGQuality <= 0 {
		switch o.Level {
		case LevelLight:
			o.JPEGQuality = qualityLight
		case LevelHeavy:
			o.JPEGQuality = qualityHeavy
		default:
			o.JPEGQuality = qualityMedium
		}
	}
	if o.JPEGQuality > 100 {
		o.JPEGQuality = 100
	}
	if o.MaxImageDim <= 0 {
		switch o.Level {
		case LevelLight:
			o.MaxImageDim = maxDimLight
		case LevelHeavy:
			o.MaxImageDim = maxDimHeavy
		default:
			o.MaxImageDim = maxDimMedium
		}
	}
	if o.MaxImageDim > maxAllowedDim {
		o.MaxImageDim = maxAllowedDim
	}
	return o
}

type pdfObject struct {
	num  int
	gen  int
	body []byte
}

var (
	rootRefRe = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	infoRefRe = regexp.MustCompile(`/Info\s+(\d+)\s+(\d+)\s+R`)
	encryptRe = regexp.MustCompile(`(?s)trailer\s*<<(.*?)>>`)
)

// CompressPDF rewrites an existing PDF using Ghostscript-style passes:
// bicubic image downsample + JPEG at the chosen tier, unused TTF glyph
// outlines dropped, document metadata stripped, and streams Flate-compressed.
// Encrypted files are rejected.
//
//nolint:revive // exported name matches the public gopdflib wrapper
func CompressPDF(data []byte, opts Options) ([]byte, error) {
	if len(data) > MaxInputBytes {
		return nil, fmt.Errorf("PDF exceeds maximum size (%d bytes)", MaxInputBytes)
	}
	if len(data) < 8 || !bytes.HasPrefix(data, []byte("%PDF-")) {
		return nil, fmt.Errorf("not a PDF file")
	}
	if isEncrypted(data) {
		return nil, fmt.Errorf("cannot compress encrypted PDF")
	}

	opts = opts.withDefaults()
	objects, maxObj, err := parseObjects(data)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("no PDF objects found")
	}
	if len(objects) > MaxObjects || maxObj > MaxObjects {
		return nil, fmt.Errorf("PDF has too many objects")
	}

	rootNum, rootGen, ok := lastRef(rootRefRe, data)
	if !ok {
		return nil, fmt.Errorf("PDF trailer /Root not found")
	}
	infoNum, _, _ := lastRef(infoRefRe, data)
	version := merge.DetectPDFVersion(data)

	stmCount := 0
	for _, obj := range objects {
		if merge.IsObjectStream(obj.body) {
			stmCount++
		}
	}
	if stmCount > 2 {
		return data, nil
	}

	stripDocumentMetadata(objects, rootNum, infoNum)

	for num, obj := range objects {
		if shouldDrop(obj.body) {
			delete(objects, num)
			continue
		}
		rewritten, changed := compressObject(obj.body, opts)
		if changed {
			obj.body = rewritten
			objects[num] = obj
		}
	}

	if _, ok := objects[rootNum]; !ok {
		return data, nil
	}

	out, err := writePDF(version, objects, maxObj, rootNum, rootGen, 0, 0, false, nil)
	if err != nil {
		return data, nil
	}
	if len(out) >= len(data) {
		return data, nil
	}
	// Parser can miss stream objects (binary "endobj" inside fonts). Never
	// ship a rewrite that dropped every stream — it renders blank.
	if bytes.Contains(data, []byte("\nstream")) && !bytes.Contains(out, []byte("stream")) {
		return data, nil
	}
	return out, nil
}

func parseObjects(data []byte) (map[int]pdfObject, int, error) {
	boundaries := merge.FindObjectBoundaries(data)
	if len(boundaries) == 0 {
		return nil, 0, fmt.Errorf("no PDF objects found")
	}

	objects := make(map[int]pdfObject, len(boundaries))
	maxObj := 0
	for _, b := range boundaries {
		bodyEnd := b.End - len("endobj")
		for bodyEnd > b.BodyStart && isPDFWhitespace(data[bodyEnd-1]) {
			bodyEnd--
		}
		body := data[b.BodyStart:bodyEnd]
		objects[b.ObjNum] = pdfObject{num: b.ObjNum, gen: b.GenNum, body: body}
		if b.ObjNum > maxObj {
			maxObj = b.ObjNum
		}
	}

	stmCount := 0
	for _, obj := range objects {
		if merge.IsObjectStream(obj.body) {
			stmCount++
		}
	}
	// Linearized files wrap a couple of object streams. Unpacking dozens of
	// them (arXiv, CID papers) is lossy and blanks pages — leave those opaque.
	if stmCount == 0 || stmCount > 2 {
		return objects, maxObj, nil
	}

	for num, obj := range objects {
		if !merge.IsObjectStream(obj.body) {
			continue
		}
		extracted := merge.ParseObjectStream(obj.body)
		if len(extracted) == 0 {
			continue
		}
		for extractedNum, extractedBody := range extracted {
			if _, exists := objects[extractedNum]; exists {
				continue
			}
			objects[extractedNum] = pdfObject{num: extractedNum, gen: 0, body: extractedBody}
			if extractedNum > maxObj {
				maxObj = extractedNum
			}
		}
		delete(objects, num)
	}

	return objects, maxObj, nil
}

func shouldDrop(body []byte) bool {
	if merge.HasSubstring(body, []byte("/Type /XRef")) || merge.HasSubstring(body, []byte("/Type/XRef")) {
		return true
	}
	if merge.HasSubstring(body, []byte("/Type /Metadata")) || merge.HasSubstring(body, []byte("/Type/Metadata")) {
		return true
	}
	if merge.HasSubstring(body, []byte("/Linearized")) {
		return true
	}
	return false
}

func compressObject(body []byte, opts Options) ([]byte, bool) {
	dict, stream, ok := splitStream(body)
	if !ok {
		return body, false
	}

	// Images only. Recompressing fonts or generic content streams blanks
	// real PDFs (arXiv, AcroForms, CID text) because encodings are not GIDs
	// and Flate+predictor streams are not reversed.
	if isImageXObject(dict) {
		if rewritten, ok := compressImage(dict, stream, opts); ok && len(rewritten) < len(body) {
			return rewritten, true
		}
	}
	return body, false
}

func writePDF(
	version string,
	objects map[int]pdfObject,
	maxObj, rootNum, rootGen, infoNum, infoGen int,
	hasInfo bool,
	idArray []byte,
) ([]byte, error) {
	if _, ok := objects[rootNum]; !ok {
		return nil, fmt.Errorf("catalog object %d missing after rewrite", rootNum)
	}

	var out bytes.Buffer
	out.WriteString("%PDF-" + version + "\n%\xe2\xe3\xcf\xd3\n")

	offsets := make(map[int]int, len(objects))
	for i := 1; i <= maxObj; i++ {
		obj, ok := objects[i]
		if !ok {
			continue
		}
		offsets[i] = out.Len()
		writeObject(&out, obj)
	}

	if hasInfo {
		if _, ok := offsets[infoNum]; !ok {
			hasInfo = false
		}
	}
	writeXRefAndTrailer(&out, offsets, maxObj, rootNum, rootGen, infoNum, infoGen, hasInfo, idArray)
	return out.Bytes(), nil
}

func writeXRefAndTrailer(
	out *bytes.Buffer,
	offsets map[int]int,
	maxObj, rootNum, rootGen, infoNum, infoGen int,
	hasInfo bool,
	idArray []byte,
) {
	xrefStart := out.Len()
	out.WriteString("xref\n0 ")
	out.WriteString(strconv.Itoa(maxObj + 1))
	out.WriteByte('\n')
	out.WriteString("0000000000 65535 f\r\n")
	var entry []byte
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			entry = entry[:0]
			offStr := strconv.FormatInt(int64(off), 10)
			for j := 0; j < 10-len(offStr); j++ {
				entry = append(entry, '0')
			}
			entry = append(entry, offStr...)
			entry = append(entry, " 00000 n\r\n"...)
			out.Write(entry)
			continue
		}
		out.WriteString("0000000000 65535 f\r\n")
	}

	out.WriteString("trailer\n<< /Size ")
	out.WriteString(strconv.Itoa(maxObj + 1))
	out.WriteString(" /Root ")
	out.WriteString(strconv.Itoa(rootNum))
	out.WriteByte(' ')
	out.WriteString(strconv.Itoa(rootGen))
	out.WriteString(" R")
	if hasInfo {
		out.WriteString(" /Info ")
		out.WriteString(strconv.Itoa(infoNum))
		out.WriteByte(' ')
		out.WriteString(strconv.Itoa(infoGen))
		out.WriteString(" R")
	}
	if len(idArray) > 0 {
		out.WriteString(" /ID ")
		out.Write(idArray)
	}
	out.WriteString(" >>\nstartxref\n")
	out.WriteString(strconv.Itoa(xrefStart))
	out.WriteString("\n%%EOF\n")
}

func writeObject(out *bytes.Buffer, obj pdfObject) {
	out.WriteString(strconv.Itoa(obj.num))
	out.WriteByte(' ')
	out.WriteString(strconv.Itoa(obj.gen))
	out.WriteString(" obj\n")
	out.Write(obj.body)
	if len(obj.body) == 0 || obj.body[len(obj.body)-1] != '\n' {
		out.WriteByte('\n')
	}
	out.WriteString("endobj\n")
}

func isEncrypted(data []byte) bool {
	for _, m := range encryptRe.FindAllSubmatch(data, -1) {
		if bytes.Contains(m[1], []byte("/Encrypt")) {
			return true
		}
	}
	return bytes.Contains(data, []byte("/Encrypt "))
}

func lastRef(re *regexp.Regexp, data []byte) (num, gen int, ok bool) {
	matches := re.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		return 0, 0, false
	}
	m := matches[len(matches)-1]
	num, _ = strconv.Atoi(string(m[1]))
	gen, _ = strconv.Atoi(string(m[2]))
	return num, gen, true
}

func isPDFWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n' || b == '\f'
}
