// Package form provides functionality for parsing XFDF and filling PDF forms.
package form

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/pdfobj"
)

var (
	reAS              = regexp.MustCompile(`/AS\s*(/(\w+)|\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	reAP              = regexp.MustCompile(`/AP\s*<<(.*?)>>`)
	reN               = regexp.MustCompile(`/N\s*<<(.*?)>>`)
	reKey             = regexp.MustCompile(`/([A-Za-z0-9_+-]+)\s*(?:/|stream|<<|\()`)
	reNName           = regexp.MustCompile(`/N\s*/([A-Za-z0-9_+-]+)`)
	reStream          = regexp.MustCompile(`(?s)stream[\r\n]+(.*?)(?:[\r\n]+endstream|endstream)`)
	reFirst           = regexp.MustCompile(`/First\s+(\d+)`)
	reAPDictForRadio  = regexp.MustCompile(`/AP\s*<<.*?/N\s*<<\s*/\s*([A-Za-z0-9_]+)\s*`)
	reRemoveAP        = regexp.MustCompile(`(?s)\s*/AP\s*<<.*?>>`)
	reASWidget        = regexp.MustCompile(`/AS\s*/\w+`)
	reValue           = regexp.MustCompile(`/V\s*(\((?:\\.|[^\\)])*\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	reAsToken         = regexp.MustCompile(`/AS\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	reParen           = regexp.MustCompile(`\(([^)]{1,200})\)`)
	reHexString       = regexp.MustCompile(`<([0-9A-Fa-f\s]{2,400})>`)
	reNameStr         = regexp.MustCompile(`/([A-Za-z0-9_+-]{1,200})`)
	reTFull           = regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	reKids            = regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	reRef             = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	reSingleKids      = regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	reVRef            = regexp.MustCompile(`/V\s*(\d+)\s+(\d+)\s+R`)
	reStreamAlt       = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	reObjStream       = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	reRect            = regexp.MustCompile(`/Rect\s*\[\s*([^\]]+)\s*\]`)
	reQ               = regexp.MustCompile(`/Q\s*(\d)`)
	reDA              = regexp.MustCompile(`/DA\s*\((.*?)\)`)
	reTf              = regexp.MustCompile(`/([\w.-]+)\s+([\d.]+)\s+Tf`)
	reVBroad          = regexp.MustCompile(`/V\s*(\((?:\\.|[^\\)])*\)|<[0-9A-Fa-f\s]+>|/[A-Za-z0-9#]+)`)
	reVParen          = regexp.MustCompile(`/V\s*\((?:\\.|[^\\)])*\)`)
	reBtnOnState      = regexp.MustCompile(`/AP\s*<<.*?/N\s*<<[^>]*?/Yes`)
	reNeedAppearances = regexp.MustCompile(`/NeedAppearances\s+(true|false)`)
	reXRefStreamObj   = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj\s*<<[^>]*?/Type\s*/XRef`)
	reAcroForm        = regexp.MustCompile(`(/AcroForm\s*<<)`)
	reRoot0           = regexp.MustCompile(`/Root\s+(\d+)\s+0\s+R`)
	reAcroFormBoth    = regexp.MustCompile(`(?s)(/AcroForm\s*<<.*?)(>>)|(/AcroForm\s+\d+\s+\d+\s+R)`)
	reAPRef           = regexp.MustCompile(`/AP\s+\d+\s+\d+\s+R`)
	reLength          = regexp.MustCompile(`/Length\s+\d+`)

	bytesSubtypeWidget      = []byte("/Subtype/Widget")
	bytesSubtypeSpaceWidget = []byte("/Subtype /Widget")
	bytesT                  = []byte("/T")
	bytesGtGt               = []byte(">>")
	bytesLtLt               = []byte("<<")
	bytesSpace              = []byte(" ")
	bytesStartxref          = []byte("startxref")
	bytesNeedAppFalse       = []byte("/NeedAppearances false")
	bytesNeedAppTrue        = []byte("/NeedAppearances true")
	bytesSpaceNeedAppFalse  = []byte(" /NeedAppearances false ")
	bytesNeedAppearances    = []byte("/NeedAppearances")
	bytesAP                 = []byte("/AP")
	bytesFieldsLb           = []byte("/Fields[")
	bytesFieldsSpLb         = []byte("/Fields [")
)

// Standard Helvetica widths for characters 32-255 (WinAnsiEncoding)
// As per the PDF 2.0 Specification - full character set for compliance
var helveticaWidths = []int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, // 32-47
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556, // 48-63
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, // 64-79
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556, // 80-95
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556, // 96-111
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584, 350, // 112-127
	556, 350, 222, 556, 333, 1000, 556, 556, 333, 1000, 667, 333, 1000, 350, 611, 350, // 128-143
	350, 222, 222, 333, 333, 350, 556, 1000, 333, 1000, 500, 333, 944, 350, 500, 667, // 144-159
	278, 333, 556, 556, 556, 556, 260, 556, 333, 737, 370, 556, 584, 333, 737, 333, // 160-175
	400, 584, 333, 333, 333, 556, 537, 278, 333, 333, 365, 556, 834, 834, 834, 611, // 176-191
	667, 667, 667, 667, 667, 667, 1000, 722, 667, 667, 667, 667, 278, 278, 278, 278, // 192-207
	722, 722, 778, 778, 778, 778, 778, 584, 778, 722, 722, 722, 722, 667, 667, 611, // 208-223
	556, 556, 556, 556, 556, 556, 889, 500, 556, 556, 556, 556, 278, 278, 278, 278, // 224-239
	556, 556, 556, 556, 556, 556, 556, 584, 611, 556, 556, 556, 556, 500, 556, 500, // 240-255
}

var helveticaWidthsStr string

func buildWidthsStr() string {
	var buf strings.Builder
	var scratch [20]byte
	buf.WriteByte('[')
	for i, w := range helveticaWidths {
		buf.Write(strconv.AppendInt(scratch[:0], int64(w), 10))
		if i < len(helveticaWidths)-1 {
			buf.WriteByte(' ')
		}
	}
	buf.WriteByte(']')
	return buf.String()
}

func init() {
	helveticaWidthsStr = buildWidthsStr()
}

func buttonOnState(dictBytes []byte) string {
	if reBtnOnState.Match(dictBytes) {
		return "/Yes"
	}
	return "/Yes"
}

// trailerHasEncrypt checks if the PDF declares document encryption: a real
// /Encrypt entry (indirect reference or inline dict) outside stream data.
// Plain "/Encrypt" text inside a content stream no longer counts.
// Owned by the pdfobj read seam.
func trailerHasEncrypt(data []byte) bool {
	return pdfobj.HasEncryptEntry(data)
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap.
// Owned by the pdfobj read seam; kept here as a thin adapter.
func parseXRefStreams(data []byte, objMap map[string][]byte) {
	pdfobj.AugmentStrMap(data, objMap)
}

// FillPDFWithXFDFAdvanced combines advanced field detection with existing value setting logic
func FillPDFWithXFDFAdvanced(pdfBytes, xfdfBytes []byte) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}

	xfdfFields, err := ParseXFDF(xfdfBytes)
	if err != nil {
		return nil, err
	}

	detectedFields, err := DetectFormFieldsAdvanced(pdfBytes)
	if err != nil {
		return nil, err
	}

	mergedFields := make(map[string]string)
	maps.Copy(mergedFields, detectedFields)
	maps.Copy(mergedFields, xfdfFields)

	// Build a synthetic XFDF from merged fields so FillPDFWithXFDF can reuse logic
	genXFDF := buildXFDF(mergedFields)
	return FillPDFWithXFDF(pdfBytes, genXFDF)
}

// FillPDFWithXFDF attempts a best-effort in-place fill of PDF form fields using XFDF data.
//
//nolint:gocyclo
func FillPDFWithXFDF(pdfBytes, xfdfBytes []byte) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	fields, err := ParseXFDF(xfdfBytes)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(pdfBytes))
	copy(out, pdfBytes)

	const (
		typeText = iota
		typeButton
		typeRadio
	)

	type job struct {
		fieldType        int
		field            string
		val              string // The correct value for this specific job
		dictStart        int
		dictEnd          int
		width, height    float64
		q                int
		fontSize         float64
		apObjNum         int
		fontObjNum       int
		fontDescObjNum   int
		fontResourceName string
		radioExportValue string
	}
	var allJobs []job

	// --- PASS 1: DISCOVERY (READ-ONLY) ---
	discoveredWidgets := make(map[string][]job)

	for fieldName, currentVal := range fields {
		dictStart, dictEnd, ok := findWidgetDictBounds(out, fieldName)
		if !ok {
			continue
		}
		dictBytes := out[dictStart:dictEnd]

		newJob := job{field: fieldName, val: currentVal, dictStart: dictStart, dictEnd: dictEnd}

		if bytes.Contains(dictBytes, []byte("/FT /Btn")) {
			if bytes.Contains(dictBytes, []byte("/Parent")) {
				newJob.fieldType = typeRadio
				if apMatch := reAPDictForRadio.FindSubmatch(dictBytes); apMatch != nil {
					newJob.radioExportValue = string(apMatch[1])
				}
			} else {
				newJob.fieldType = typeButton
			}
		} else if bytes.Contains(dictBytes, []byte("/FT /Tx")) {
			newJob.fieldType = typeText
			rectMatch := reRect.FindSubmatch(dictBytes)
			if rectMatch == nil {
				continue
			}
			coords := strings.Fields(string(rectMatch[1]))
			if len(coords) < 4 {
				continue
			}
			llx, _ := strconv.ParseFloat(coords[0], 64)
			lly, _ := strconv.ParseFloat(coords[1], 64)
			urx, _ := strconv.ParseFloat(coords[2], 64)
			ury, _ := strconv.ParseFloat(coords[3], 64)
			newJob.width, newJob.height = urx-llx, ury-lly

			if qMatch := reQ.FindSubmatch(dictBytes); qMatch != nil {
				newJob.q, _ = strconv.Atoi(string(qMatch[1]))
			}
			newJob.fontSize, newJob.fontResourceName = 12.0, "Helv" // Default to Helv
			if daMatch := reDA.FindSubmatch(dictBytes); daMatch != nil {
				if tfMatch := reTf.FindStringSubmatch(string(daMatch[1])); len(tfMatch) > 2 {
					// NOTE: This logic assumes standard PDF fonts. If the original font was embedded,
					// we are replacing it with a standard one (Helvetica). This is usually fine.
					newJob.fontResourceName = "Helv" // Force Helvetica for consistency
					if fs, err := strconv.ParseFloat(tfMatch[2], 64); err == nil {
						newJob.fontSize = fs
					}
				}
			}
		}
		discoveredWidgets[fieldName] = append(discoveredWidgets[fieldName], newJob)
	}
	for _, jobs := range discoveredWidgets {
		allJobs = append(allJobs, jobs...)
	}

	objStmChanged := false
	if bytes.Contains(out, []byte("/ObjStm")) {
		if filled, changed, err := fillXFDFInObjectStreams(out, fields); err != nil {
			return nil, err
		} else if changed {
			out = filled
			objStmChanged = true
		}
	}

	// Object-stream-only fills keep the original xref/trailer; rebuilding breaks PDF 1.5+ xref streams.
	if len(allJobs) == 0 {
		return repairStartxref(out), nil
	}

	if !objStmChanged {

		// --- PASS 2: MODIFICATION (WRITE-ONLY, IN REVERSE) ---
		sort.Slice(allJobs, func(i, j int) bool { return allJobs[i].dictStart > allJobs[j].dictStart })

		radioGroups := make(map[string]string)
		for _, job := range allJobs {
			if job.fieldType == typeRadio {
				radioGroups[job.field] = job.val
			}
		}
		for fieldName, value := range radioGroups {
			tKey := []byte("/T (" + fieldName + ")")
			tIdx := bytes.Index(out, tKey)
			if tIdx < 0 {
				continue
			}
			dictStart := bytes.LastIndex(out[:tIdx], []byte("<<"))
			if dictStart < 0 {
				continue
			}
			dictEnd := bytes.Index(out[dictStart:], []byte(">>"))
			if dictEnd < 0 {
				continue
			}
			dictEnd += dictStart + 2
			if !bytes.Contains(out[dictStart:dictEnd], []byte("/Kids")) || bytes.Contains(out[dictStart:dictEnd], bytesSubtypeSpaceWidget) {
				continue
			}
			dictBytes := out[dictStart:dictEnd]
			newV := make([]byte, 0, len(value)+4)
			newV = append(newV, "/V /"...)
			newV = append(newV, value...)
			var newDictBytes []byte
			if reVBroad.Match(dictBytes) {
				newDictBytes = reVBroad.ReplaceAll(dictBytes, newV)
			} else {
				newDictBytes = bytes.Replace(dictBytes, bytesGtGt, append(bytesSpace, append(newV, bytesGtGt...)...), 1)
			}
			out = append(out[:dictStart], append(newDictBytes, out[dictEnd:]...)...)
		}

		for _, job := range allJobs {
			dictStart, dictEnd, ok := findWidgetDictBounds(out, job.field)
			if !ok {
				continue
			}
			dictBytes := out[dictStart:dictEnd]
			var newDictBytes []byte
			switch job.fieldType {
			case typeText:
				esc := escapePDFString(job.val)
				newV := make([]byte, 0, len(esc)+6)
				newV = append(newV, "/V ("...)
				newV = append(newV, esc...)
				newV = append(newV, ')')
				if reVParen.Match(dictBytes) {
					newDictBytes = reVParen.ReplaceAll(dictBytes, newV)
				} else {
					newDictBytes = bytes.Replace(dictBytes, bytesGtGt, append(bytesSpace, append(newV, bytesGtGt...)...), 1)
				}
				newDictBytes = reRemoveAP.ReplaceAll(newDictBytes, bytesSpace)
			case typeButton, typeRadio:
				newState := []byte("/Off")
				if job.fieldType == typeButton && (strings.EqualFold(job.val, "yes") || strings.EqualFold(job.val, "on")) {
					newState = []byte(buttonOnState(dictBytes))
				} else if job.fieldType == typeRadio && job.radioExportValue == job.val {
					newState = append([]byte("/"), job.radioExportValue...)
				}
				newAS := append([]byte("/AS "), newState...)
				newVBtn := append([]byte("/V "), newState...)
				newDictBytes = dictBytes
				if reASWidget.Match(newDictBytes) {
					newDictBytes = reASWidget.ReplaceAll(newDictBytes, newAS)
				} else {
					newDictBytes = bytes.Replace(newDictBytes, bytesGtGt, append(bytesSpace, append(newAS, bytesGtGt...)...), 1)
				}
				if reValue.Match(newDictBytes) {
					newDictBytes = reValue.ReplaceAll(newDictBytes, newVBtn)
				} else {
					newDictBytes = bytes.Replace(newDictBytes, bytesGtGt, append(bytesSpace, append(newVBtn, bytesGtGt...)...), 1)
				}
			}
			if newDictBytes != nil {
				out = append(out[:dictStart], append(newDictBytes, out[dictEnd:]...)...)
			}
		}
	}

	// --- PASS 3: NEW OBJECT GENERATION ---
	objRe := regexp.MustCompile(`(\d+)\s+0\s+obj`)
	allObjMatches := objRe.FindAllSubmatchIndex(out, -1)
	highest := 0
	for _, m := range allObjMatches {
		if n, err := strconv.Atoi(string(out[m[2]:m[3]])); err == nil && n > highest {
			highest = n
		}
	}
	nextObj := highest + 1
	if sx := bytes.LastIndex(out, bytesStartxref); sx >= 0 {
		out = out[:sx]
	}
	var textJobs []*job
	for i := range allJobs {
		if allJobs[i].fieldType == typeText {
			textJobs = append(textJobs, &allJobs[i])
		}
	}
	sort.Slice(textJobs, func(i, j int) bool { return textJobs[i].dictStart > textJobs[j].dictStart })

	for _, job := range textJobs {
		job.fontDescObjNum = nextObj
		nextObj++
		job.fontObjNum = nextObj
		nextObj++
		job.apObjNum = nextObj
		nextObj++
		dictStart, dictEnd, ok := findWidgetDictBounds(out, job.field)
		if !ok {
			continue
		}
		dictBytes := out[dictStart:dictEnd]
		dictBytes = reRemoveAP.ReplaceAll(dictBytes, bytesSpace)
		var apRefScratch [30]byte
		b := append(apRefScratch[:0], " /AP<</N "...)
		b = strconv.AppendInt(b, int64(job.apObjNum), 10)
		b = append(b, " 0 R>>"...)
		apRef := bytes.Clone(b)
		var newDict bytes.Buffer
		newDict.Grow(len(dictBytes) + len(apRef))
		newDict.Write(dictBytes[:len(dictBytes)-2])
		newDict.Write(apRef)
		newDict.Write(dictBytes[len(dictBytes)-2:])
		out = append(out[:dictStart], append(newDict.Bytes(), out[dictEnd:]...)...)
	}

	if reNeedAppearances.Match(out) {
		out = reNeedAppearances.ReplaceAll(out, bytesNeedAppFalse)
	} else {
		if loc := reAcroForm.FindIndex(out); loc != nil {
			insertPos := loc[1]
			insertContent := bytesSpaceNeedAppFalse
			newOut := make([]byte, 0, len(out)+len(insertContent))
			newOut = append(newOut, out[:insertPos]...)
			newOut = append(newOut, insertContent...)
			newOut = append(newOut, out[insertPos:]...)
			out = newOut
		}
	}

	var buf bytes.Buffer
	var scratch [40]byte
	for _, job := range textJobs {
		streamText := escapePDFString(job.val)
		var tx float64
		textWidth := float64(len(job.val)) * job.fontSize * 0.55
		switch job.q {
		case 1:
			tx = (job.width - textWidth) / 2
		case 2:
			tx = job.width - textWidth - 3
		default:
			tx = 3
		}
		if tx < 3 {
			tx = 3
		}
		y := (job.height-job.fontSize)/2 + 1.5
		if y < 2 {
			y = 2
		}

		buf.Reset()
		buf.WriteString("q\nBT\n/F1 ")
		buf.Write(strconv.AppendFloat(scratch[:0], job.fontSize, 'f', 2, 64))
		buf.WriteString(" Tf\n0 g\n")
		buf.Write(strconv.AppendFloat(scratch[:0], tx, 'f', 2, 64))
		buf.WriteByte(' ')
		buf.Write(strconv.AppendFloat(scratch[:0], y, 'f', 2, 64))
		buf.WriteString(" Td\n(")
		buf.WriteString(streamText)
		buf.WriteString(") Tj\nET\nQ")
		streamBytes := buf.Bytes()

		buf.Reset()
		buf.WriteByte('\n')
		buf.Write(strconv.AppendInt(scratch[:0], int64(job.fontDescObjNum), 10))
		buf.WriteString(" 0 obj\n<</Type/FontDescriptor/FontName/")
		buf.WriteString(job.fontResourceName)
		buf.WriteString("/Flags 32/FontBBox[-558 -225 1000 931]/ItalicAngle 0/Ascent 905/Descent -212/CapHeight 905/StemV 88>>\nendobj\n")
		out = append(out, buf.Bytes()...)

		buf.Reset()
		buf.WriteByte('\n')
		buf.Write(strconv.AppendInt(scratch[:0], int64(job.fontObjNum), 10))
		buf.WriteString(" 0 obj\n<</Type/Font/Subtype/Type1/BaseFont/")
		buf.WriteString(job.fontResourceName)
		buf.WriteString("/Encoding/WinAnsiEncoding/FirstChar 32/LastChar 255/Widths ")
		buf.WriteString(helveticaWidthsStr)
		buf.WriteString("/FontDescriptor ")
		buf.Write(strconv.AppendInt(scratch[:0], int64(job.fontDescObjNum), 10))
		buf.WriteString(" 0 R>>\nendobj\n")
		out = append(out, buf.Bytes()...)

		var compBuf bytes.Buffer
		zw, _ := zlib.NewWriterLevel(&compBuf, zlib.BestCompression)
		if _, err := zw.Write(streamBytes); err != nil {
			return nil, fmt.Errorf("compression write failed: %w", err)
		}
		if err := zw.Close(); err != nil {
			return nil, fmt.Errorf("compression close failed: %w", err)
		}
		comp := compBuf.Bytes()

		buf.Reset()
		buf.WriteByte('\n')
		buf.Write(strconv.AppendInt(scratch[:0], int64(job.apObjNum), 10))
		buf.WriteString(" 0 obj\n<</Type/XObject/Subtype/Form/FormType 1/BBox[0 0 ")
		buf.Write(strconv.AppendFloat(scratch[:0], job.width, 'f', 2, 64))
		buf.WriteByte(' ')
		buf.Write(strconv.AppendFloat(scratch[:0], job.height, 'f', 2, 64))
		buf.WriteString("]/Resources<</Font<</F1 ")
		buf.Write(strconv.AppendInt(scratch[:0], int64(job.fontObjNum), 10))
		buf.WriteString(" 0 R>>/ProcSet[/PDF/Text]>>/Filter/FlateDecode/Length ")
		buf.Write(strconv.AppendInt(scratch[:0], int64(len(comp)), 10))
		buf.WriteString(">>\nstream\n")
		buf.Write(comp)
		buf.WriteString("\nendstream\nendobj\n")
		out = append(out, buf.Bytes()...)
	}

	// Apply NeedAppearances before writing xref/trailer so byte offsets stay consistent.
	out = setAcroFormNeedAppearancesTrue(out)

	objMatches := objRe.FindAllSubmatchIndex(out, -1)
	offsets := make(map[int]int)
	maxObj := 0
	for _, m := range objMatches {
		num, _ := strconv.Atoi(string(out[m[2]:m[3]]))
		offsets[num] = m[0]
		if num > maxObj {
			maxObj = num
		}
	}
	xrefStart := len(out)
	var xrefBuf bytes.Buffer
	var xrefPad [20]byte
	xrefBuf.WriteString("xref\n0 ")
	xrefBuf.Write(strconv.AppendInt(xrefPad[:0], int64(maxObj+1), 10))
	xrefBuf.WriteString("\n")
	xrefBuf.WriteString("0000000000 65535 f \r\n")
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			b := strconv.AppendInt(xrefPad[:0], int64(off), 10)
			for j := len(b); j < 10; j++ {
				xrefBuf.WriteByte('0')
			}
			xrefBuf.Write(b)
			xrefBuf.WriteString(" 00000 n \r\n")
		} else {
			xrefBuf.WriteString("0000000000 65535 f \r\n")
		}
	}
	root := 1
	if rm := reRoot0.FindSubmatch(pdfBytes); len(rm) > 1 {
		if r, err := strconv.Atoi(string(rm[1])); err == nil {
			root = r
		}
	}
	var trailerBuf bytes.Buffer
	trailerBuf.WriteString("trailer\n<</Size ")
	trailerBuf.Write(strconv.AppendInt(scratch[:0], int64(maxObj+1), 10))
	trailerBuf.WriteString("/Root ")
	trailerBuf.Write(strconv.AppendInt(scratch[:0], int64(root), 10))
	trailerBuf.WriteString(" 0 R>>\nstartxref\n")
	trailerBuf.Write(strconv.AppendInt(scratch[:0], int64(xrefStart), 10))
	trailerBuf.WriteString("\n%%%%EOF\n")
	out = append(out, xrefBuf.Bytes()...)
	out = append(out, trailerBuf.Bytes()...)

	return out, nil
}

// repairStartxref rewrites the trailer offset after in-place edits shift object positions.
func repairStartxref(out []byte) []byte {
	sxIdx := bytes.LastIndex(out, bytesStartxref)
	if sxIdx < 0 {
		return out
	}

	body := out[:sxIdx]
	xrefOffset := -1
	if matches := reXRefStreamObj.FindAllSubmatchIndex(body, -1); len(matches) > 0 {
		xrefOffset = matches[len(matches)-1][0]
	} else if idx := bytes.LastIndex(body, []byte("\nxref\n")); idx >= 0 {
		xrefOffset = idx + 1
	} else if idx := bytes.LastIndex(body, []byte("\nxref\r\n")); idx >= 0 {
		xrefOffset = idx + 1
	}
	if xrefOffset < 0 {
		return out
	}

	after := out[sxIdx+len("startxref"):]
	after = bytes.TrimLeft(after, "\r\n")
	_, after0, ok := bytes.Cut(after, []byte{'\n'})
	if !ok {
		return out
	}
	rest := after0

	var rebuilt bytes.Buffer
	rebuilt.Grow(len(out) + 16)
	rebuilt.Write(out[:sxIdx+len("startxref")])
	rebuilt.WriteByte('\n')
	rebuilt.Write(strconv.AppendInt(make([]byte, 0, 12), int64(xrefOffset), 10))
	rebuilt.WriteByte('\n')
	rebuilt.Write(rest)
	return rebuilt.Bytes()
}

// setAcroFormNeedAppearancesTrue forces viewers to regenerate field appearances on open.
// Must run before xref/trailer emission; any later byte insertion invalidates startxref.
func setAcroFormNeedAppearancesTrue(out []byte) []byte {
	acroMatch := reAcroFormBoth.FindSubmatch(out)
	if acroMatch == nil {
		return out
	}
	if acroMatch[1] != nil {
		dictPart := acroMatch[1]
		if reNeedAppearances.Match(dictPart) {
			return bytes.Replace(out, dictPart, reNeedAppearances.ReplaceAll(dictPart, bytesNeedAppTrue), 1)
		}
		newDict := make([]byte, len(dictPart)+len(" /NeedAppearances true "))
		copy(newDict, dictPart)
		copy(newDict[len(dictPart):], " /NeedAppearances true ")
		return bytes.Replace(out, dictPart, newDict, 1)
	}
	if acroMatch[3] != nil {
		refFull := string(acroMatch[3])
		if rm := reRef.FindStringSubmatch(refFull); rm != nil {
			objRe := regexp.MustCompile(`(?s)\b` + rm[1] + `\s+` + rm[2] + `\s+obj(.*?)endobj`)
			if objM := objRe.FindSubmatch(out); objM != nil {
				objBody := objM[1]
				if reNeedAppearances.Match(objBody) {
					return bytes.Replace(out, objBody, reNeedAppearances.ReplaceAll(objBody, bytesNeedAppTrue), 1)
				}
				insertPos := bytes.LastIndex(objBody, bytesGtGt)
				if insertPos >= 0 {
					var newBody bytes.Buffer
					newBody.Write(objBody[:insertPos])
					newBody.WriteString(" /NeedAppearances true ")
					newBody.Write(objBody[insertPos:])
					return bytes.Replace(out, objBody, newBody.Bytes(), 1)
				}
			}
		}
	}
	return out
}

// FlattenPDFBytes flattens form fields from the provided PDF bytes and returns flattened PDF bytes.
// This is a local stub so callers in this package can invoke flattening. Replace implementation
// to call the shared flattener when available.
func FlattenPDFBytes(pdfBytes []byte) ([]byte, error) {
	// TODO: call the canonical flattener (e.g., scripts.FlattenPDFBytes) once moved into an importable package.
	// For now, return the input bytes unchanged.
	// return scripts.FlattenPDFBytes(pdfBytes)
	return pdfBytes, nil
}

func fillXFDFInObjectStreams(pdfBytes []byte, fields map[string]string) ([]byte, bool, error) {
	out := make([]byte, len(pdfBytes))
	copy(out, pdfBytes)

	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	matches := objRe.FindAllSubmatchIndex(out, -1)
	if len(matches) == 0 {
		return out, false, nil
	}

	changedAny := false
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		bodyStart, bodyEnd := m[6], m[7]
		body := out[bodyStart:bodyEnd]
		if !bytes.Contains(body, []byte("/ObjStm")) {
			continue
		}

		newBody, changed, err := fillXFDFInObjStmBody(body, fields)
		if err != nil {
			return nil, false, err
		}
		if !changed {
			continue
		}

		changedAny = true
		out = append(out[:bodyStart], append(newBody, out[bodyEnd:]...)...)
	}

	return out, changedAny, nil
}

//nolint:gocyclo
func fillXFDFInObjStmBody(body []byte, fields map[string]string) ([]byte, bool, error) {
	streamRe := regexp.MustCompile(`(?s)stream[\r\n]+(.*?)(?:[\r\n]+endstream|endstream)`)
	sm := streamRe.FindSubmatchIndex(body)
	if sm == nil {
		return body, false, nil
	}

	streamBytes := body[sm[2]:sm[3]]
	var decoded []byte
	if d, err := tryZlibDecompress(streamBytes); err == nil {
		decoded = d
	} else if d, err := tryFlateDecompress(streamBytes); err == nil {
		decoded = d
	} else {
		return body, false, nil
	}

	fm := reFirst.FindSubmatch(body)
	if fm == nil {
		return body, false, nil
	}
	first, err := strconv.Atoi(string(fm[1]))
	if err != nil || first <= 0 || first > len(decoded) {
		return body, false, nil
	}

	header := strings.TrimSpace(string(decoded[:first]))
	parts := strings.Fields(header)
	if len(parts) < 2 || len(parts)%2 != 0 {
		return body, false, nil
	}

	type objMember struct {
		objNum  int
		offset  int
		content []byte
	}

	content := decoded[first:]
	members := make([]objMember, 0, len(parts)/2)
	for i := 0; i < len(parts); i += 2 {
		num, numErr := strconv.Atoi(parts[i])
		off, offErr := strconv.Atoi(parts[i+1])
		if numErr != nil || offErr != nil {
			return body, false, nil
		}
		members = append(members, objMember{objNum: num, offset: off})
	}

	kidsToRemoveAP := make(map[int]bool)
	nameRe := regexp.MustCompile(`/T\s*(?:\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	singleKidsRe := regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	for i := range members {
		start := members[i].offset
		end := len(content)
		if i+1 < len(members) {
			end = members[i+1].offset
		}
		objContent := bytes.TrimSpace(content[start:end])

		if nameMatch := nameRe.FindSubmatch(objContent); nameMatch != nil {
			var fieldName string
			if len(nameMatch[1]) > 0 {
				fieldName = string(nameMatch[1])
			} else if len(nameMatch[2]) > 0 {
				fieldName = decodeHexString(string(nameMatch[2]))
			}
			fieldName = strings.TrimSpace(fieldName)

			if _, ok := fields[fieldName]; ok {
				if bytes.Contains(objContent, []byte("/FT /Tx")) || bytes.Contains(objContent, []byte("/FT/Tx")) {
					kidsToRemoveAP[members[i].objNum] = true
					if m := kidsRe.FindSubmatch(objContent); m != nil {
						for _, r := range refRe.FindAllSubmatch(m[1], -1) {
							kidNum, _ := strconv.Atoi(string(r[1]))
							kidsToRemoveAP[kidNum] = true
						}
					}
					if m := singleKidsRe.FindSubmatch(objContent); m != nil {
						kidNum, _ := strconv.Atoi(string(m[1]))
						kidsToRemoveAP[kidNum] = true
					}
				}
			}
		}
	}

	changedAny := false
	for i := range members {
		start := members[i].offset
		end := len(content)
		if i+1 < len(members) {
			end = members[i+1].offset
		}
		if start < 0 || start > len(content) || end < start || end > len(content) {
			return body, false, nil
		}

		objContent := bytes.TrimSpace(content[start:end])
		updated, changed := updateObjStmFieldValue(objContent, fields)

		if kidsToRemoveAP[members[i].objNum] {
			// Clean any standalone indirect APs
			if reAPRef.Match(updated) {
				updated = reAPRef.ReplaceAll(updated, bytesSpace)
				changed = true
			}
			// Manual removal of /AP dictionary to handle nested <<>>
			apIdx := bytes.Index(updated, bytesAP)
			if apIdx >= 0 {
				afterAP := updated[apIdx+3:]
				trimmedAfter := bytes.TrimSpace(afterAP)
				if bytes.HasPrefix(trimmedAfter, bytesLtLt) {
					// We need to find matching >>
					depth := 0
					endIdx := -1
					for j := 0; j < len(trimmedAfter)-1; j++ {
						if trimmedAfter[j] == '<' && trimmedAfter[j+1] == '<' {
							depth++
							j++
						} else if trimmedAfter[j] == '>' && trimmedAfter[j+1] == '>' {
							depth--
							j++
							if depth == 0 {
								endIdx = j
								break
							}
						}
					}
					if endIdx != -1 {
						// Remove from /AP through the matching >>
						startRemove := apIdx
						endRemove := apIdx + 3 + (len(afterAP) - len(trimmedAfter)) + endIdx + 1

						var newBody bytes.Buffer
						newBody.Write(updated[:startRemove])
						newBody.WriteByte(' ') // replace with space
						newBody.Write(updated[endRemove:])
						updated = newBody.Bytes()
						changed = true
					}
				}
			}
		}

		// Pass 3: If this object is AcroForm globally inject NeedAppearances.
		// Usually identified by /Fields or /DA or /SigFlags coupled with being a catalog reference...
		if bytes.Contains(updated, bytesFieldsLb) || bytes.Contains(updated, bytesFieldsSpLb) {
			// Basic AcroForm object identification
			if !bytes.Contains(updated, bytesNeedAppearances) {
				insertPos := bytes.LastIndex(updated, bytesGtGt)
				if insertPos >= 0 {
					var newBody bytes.Buffer
					newBody.Write(updated[:insertPos])
					newBody.WriteString(" /NeedAppearances true ")
					newBody.Write(updated[insertPos:])
					updated = newBody.Bytes()
					changed = true
				}
			}
		}

		if changed {
			changedAny = true
		}
		members[i].content = updated
	}

	if !changedAny {
		return body, false, nil
	}

	var headerBuilder bytes.Buffer
	var contentBuilder bytes.Buffer
	var memScratch [20]byte
	currentOffset := 0
	for i, member := range members {
		headerBuilder.Write(strconv.AppendInt(memScratch[:0], int64(member.objNum), 10))
		headerBuilder.WriteByte(' ')
		headerBuilder.Write(strconv.AppendInt(memScratch[:0], int64(currentOffset), 10))
		if i != len(members)-1 {
			headerBuilder.WriteByte(' ')
		}

		contentBuilder.Write(member.content)
		if i != len(members)-1 {
			contentBuilder.WriteByte(' ')
			currentOffset += len(member.content) + 1
		} else {
			currentOffset += len(member.content)
		}
	}

	newHeaderBytes := headerBuilder.Bytes()
	newFirst := len(newHeaderBytes) + 1
	contentBytes := contentBuilder.Bytes()
	newDecoded := make([]byte, 0, len(newHeaderBytes)+1+len(contentBytes))
	newDecoded = append(newDecoded, bytes.Clone(newHeaderBytes)...)
	newDecoded = append(newDecoded, ' ')
	newDecoded = append(newDecoded, contentBytes...)

	var compressedBuf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&compressedBuf, zlib.BestCompression)
	if err != nil {
		return nil, false, err
	}
	if _, err := zw.Write(newDecoded); err != nil {
		return nil, false, err
	}
	if err := zw.Close(); err != nil {
		return nil, false, err
	}
	compressed := compressedBuf.Bytes()

	dictPart := body[:sm[0]]
	suffix := body[sm[1]:]
	var firstRepl [30]byte
	fb := strconv.AppendInt(append(firstRepl[:0], "/First "...), int64(newFirst), 10)
	newDict := reFirst.ReplaceAll(dictPart, fb)
	if reLength.Match(newDict) {
		var lenRepl [30]byte
		lb := strconv.AppendInt(append(lenRepl[:0], "/Length "...), int64(len(compressed)), 10)
		newDict = reLength.ReplaceAll(newDict, lb)
	}

	var rebuilt bytes.Buffer
	rebuilt.Write(newDict)
	rebuilt.WriteString("stream\n")
	rebuilt.Write(compressed)
	rebuilt.WriteString("\nendstream")
	rebuilt.Write(suffix)

	return rebuilt.Bytes(), true, nil
}

func updateObjStmFieldValue(objContent []byte, fields map[string]string) ([]byte, bool) {
	nameRe := regexp.MustCompile(`/T\s*(?:\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	nameMatch := nameRe.FindSubmatch(objContent)
	if nameMatch == nil {
		return objContent, false
	}

	var fieldName string
	if len(nameMatch[1]) > 0 {
		fieldName = string(nameMatch[1])
	} else if len(nameMatch[2]) > 0 {
		fieldName = decodeHexString(string(nameMatch[2]))
	}
	fieldName = strings.TrimSpace(fieldName)

	value, ok := fields[fieldName]
	if !ok {
		return objContent, false
	}

	updated := make([]byte, len(objContent))
	copy(updated, objContent)

	if bytes.Contains(updated, []byte("/FT /Tx")) || bytes.Contains(updated, []byte("/FT/Tx")) {
		replacement := []byte("/V (" + escapePDFString(value) + ")")
		newUpdated, changed := replaceOrInsertPDFEntry(updated, `/V\s*\((?:\\.|[^\\)])*\)|/V\s*/[^\s/>]+|/V\s*<[0-9A-Fa-f\s]+>`, replacement)
		return newUpdated, changed
	}

	if bytes.Contains(updated, []byte("/FT /Btn")) || bytes.Contains(updated, []byte("/FT/Btn")) {
		state := xfdfValueToPDFName(value)
		newUpdated, changedV := replaceOrInsertPDFEntry(updated, `/V\s*/[^\s/>]+|/V\s*\((?:\\.|[^\\)])*\)|/V\s*<[0-9A-Fa-f\s]+>`, []byte("/V /"+state))
		newUpdated, changedAS := replaceOrInsertPDFEntry(newUpdated, `/AS\s*/[^\s/>]+|/AS\s*\((?:\\.|[^\\)])*\)|/AS\s*<[0-9A-Fa-f\s]+>`, []byte("/AS /"+state))
		return newUpdated, changedV || changedAS
	}

	return objContent, false
}

func replaceOrInsertPDFEntry(dict []byte, pattern string, replacement []byte) ([]byte, bool) {
	re := regexp.MustCompile(pattern)
	if re.Match(dict) {
		newDict := re.ReplaceAll(dict, replacement)
		return newDict, len(newDict) != len(dict) || !bytes.Equal(newDict, dict)
	}

	insertPos := bytes.LastIndex(dict, bytesGtGt)
	if insertPos < 0 {
		// If '>>' is not found, append to the end.
		// Ensure the original 'dict' is not used after 'append' if it reallocates.
		newDict := append(dict, ' ') //nolint:gocritic
		newDict = append(newDict, replacement...)
		return newDict, true
	}

	var out bytes.Buffer
	out.Write(dict[:insertPos])
	out.WriteByte(' ')
	out.Write(replacement)
	out.Write(dict[insertPos:])
	return out.Bytes(), true
}

func xfdfValueToPDFName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "Off"
	}
	if strings.EqualFold(trimmed, "off") || strings.EqualFold(trimmed, "false") || strings.EqualFold(trimmed, "no") || trimmed == "0" {
		return "Off"
	}

	var b strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	token := b.String()
	if token == "" {
		return "Yes"
	}
	return token
}
