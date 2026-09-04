// Package pdfobj owns the PDF read seam: object boundaries, trailer and
// startxref scanning, xref-table and xref-stream parsing, object-stream
// expansion, and capped inflation. All other packages parse PDFs through
// this seam instead of duplicating scanners.
package pdfobj

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// MaxXRefEntryWidth caps the total byte width of one cross-reference stream
// entry (/W array sum). Field widths come from the file, so they must be
// validated before slicing: W=[0 0 0] would make the entry loop never
// advance (hang) and negative widths panic on slicing.
const MaxXRefEntryWidth = 32

// MaxInflateBytes caps a single decompressed stream (48 MiB) so a crafted
// Flate stream cannot act as a zip bomb.
const MaxInflateBytes = 48 << 20

// InflateCapped decompresses zlib-wrapped (rawFlate=false) or raw flate data,
// rejecting outputs over MaxInflateBytes.
func InflateCapped(b []byte, rawFlate bool) ([]byte, error) {
	var r io.ReadCloser
	if rawFlate {
		r = flate.NewReader(bytes.NewReader(b))
	} else {
		var err error
		r, err = zlib.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		_ = r.Close()
	}()
	var out bytes.Buffer
	n, err := io.Copy(&out, io.LimitReader(r, int64(MaxInflateBytes)+1))
	if err != nil {
		return nil, err
	}
	if n > int64(MaxInflateBytes) {
		return nil, fmt.Errorf("decompressed stream exceeds %d bytes", MaxInflateBytes)
	}
	return out.Bytes(), nil
}

// ValidXRefWidths checks /W field widths from a cross-reference stream.
// It returns the entry width and true when the widths are safe to slice
// with: all non-negative, total in (0, MaxXRefEntryWidth].
func ValidXRefWidths(w0, w1, w2 int) (total int, ok bool) {
	if w0 < 0 || w1 < 0 || w2 < 0 {
		return 0, false
	}
	total = w0 + w1 + w2
	if total <= 0 || total > MaxXRefEntryWidth {
		return 0, false
	}
	return total, true
}

// InflateZlib is InflateCapped with zlib wrapping. InflateFlate is the raw
// flate variant. Both exist so callers can delegate their near-duplicate
// tryZlibDecompress / tryFlateDecompress wrappers here.
func InflateZlib(b []byte) ([]byte, error) { return InflateCapped(b, false) }

// InflateFlate is InflateCapped with raw flate data.
func InflateFlate(b []byte) ([]byte, error) { return InflateCapped(b, true) }

// DecompressAny tries zlib, then raw flate, returning the first success.
// It mirrors the try-zlib-then-flate fallback chains in compress, redact,
// form, and pdf helpers.
func DecompressAny(b []byte) ([]byte, error) {
	if out, err := InflateCapped(b, false); err == nil {
		return out, nil
	}
	return InflateCapped(b, true)
}

// DecompressFlateResult decompresses zlib data and returns nil (not an
// error) when the input is not valid zlib, matching the legacy merge
// decompressFlate behavior used during object-stream expansion.
func DecompressFlateResult(data []byte) []byte {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer func() {
		_ = reader.Close()
	}()

	const maxInflateBytes = 48 << 20
	var out bytes.Buffer
	n, err := io.Copy(&out, io.LimitReader(reader, maxInflateBytes+1))
	if err != nil || n > maxInflateBytes {
		return nil
	}

	return out.Bytes()
}

// ObjectBoundary represents the position of a PDF object in the file.
type ObjectBoundary struct {
	ObjNum    int
	GenNum    int
	Start     int
	BodyStart int
	End       int
}

// DetectVersion extracts the PDF version from the header (e.g. "1.4", "1.7", "2.0").
func DetectVersion(data []byte) string {
	versionRe := regexp.MustCompile(`%PDF-(\d+\.\d+)`)
	if m := versionRe.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return "1.7"
}

// CompareVersions compares two PDF version strings.
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
func CompareVersions(v1, v2 string) int {
	parse := func(v string) (int, int) {
		parts := bytes.Split([]byte(v), []byte("."))
		major, _ := strconv.Atoi(string(parts[0]))
		minor := 0
		if len(parts) > 1 {
			minor, _ = strconv.Atoi(string(parts[1]))
		}
		return major, minor
	}

	maj1, min1 := parse(v1)
	maj2, min2 := parse(v2)

	if maj1 != maj2 {
		if maj1 > maj2 {
			return 1
		}
		return -1
	}
	if min1 > min2 {
		return 1
	} else if min1 < min2 {
		return -1
	}
	return 0
}

// FindObjectBoundaries finds all PDF objects in the data.
func FindObjectBoundaries(data []byte) []ObjectBoundary {
	var results = make([]ObjectBoundary, 0, len(data)/200)
	objStartRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)

	pos := 0
	for pos < len(data) {
		loc := objStartRe.FindSubmatchIndex(data[pos:])
		if loc == nil {
			break
		}

		start := pos + loc[0]
		bodyStart := pos + loc[1]

		objNum, _ := strconv.Atoi(string(data[pos+loc[2] : pos+loc[3]]))
		genNum, _ := strconv.Atoi(string(data[pos+loc[4] : pos+loc[5]]))

		endPos := FindEndObj(data, bodyStart)
		if endPos == -1 {
			pos = start + 1
			continue
		}

		results = append(results, ObjectBoundary{
			ObjNum:    objNum,
			GenNum:    genNum,
			Start:     start,
			BodyStart: bodyStart,
			End:       endPos,
		})

		pos = endPos
	}

	return results
}

// FindEndObj finds the position right after "endobj" starting from pos.
func FindEndObj(data []byte, pos int) int {
	i := pos
	n := len(data)

	for i < n {
		for i < n && IsWhitespace(data[i]) {
			i++
		}
		if i >= n {
			break
		}

		if i+6 <= n && string(data[i:i+6]) == "endobj" {
			return i + 6
		}

		if i+6 <= n && string(data[i:i+6]) == "stream" {
			if i+6 < n {
				b := data[i+6]
				if b == '\r' || b == '\n' {
					searchPos := i + 6
					for {
						idx := bytes.Index(data[searchPos:], []byte("endstream"))
						if idx == -1 {
							return -1
						}

						matchPos := searchPos + idx

						validEnd := false
						if matchPos > 0 {
							switch data[matchPos-1] {
							case '\n', '\r':
								validEnd = true
							}
						}

						if validEnd {
							i = matchPos + 9
							break
						}

						searchPos = matchPos + 1
					}
					continue
				}
			}
		}

		switch {
		case data[i] == '(':
			i = SkipStringLiteral(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] != '<':
			i = SkipHexString(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] == '<':
			i = SkipDictionary(data, i)
		case data[i] == '[':
			i = SkipArray(data, i)
		default:
			i++
		}
	}

	return -1
}

// SkipStringLiteral skips a PDF string literal (...) handling escapes and nested parens.
func SkipStringLiteral(data []byte, pos int) int {
	if pos >= len(data) || data[pos] != '(' {
		return pos + 1
	}
	i := pos + 1
	depth := 1
	for i < len(data) && depth > 0 {
		if data[i] == '\\' && i+1 < len(data) {
			i += 2
			continue
		}
		switch data[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		i++
	}
	return i
}

// SkipHexString skips a PDF hex string <...>.
func SkipHexString(data []byte, pos int) int {
	if pos >= len(data) || data[pos] != '<' {
		return pos + 1
	}
	i := pos + 1
	for i < len(data) && data[i] != '>' {
		i++
	}
	return i + 1
}

// SkipDictionary skips a PDF dictionary <<...>>.
func SkipDictionary(data []byte, pos int) int {
	if pos+1 >= len(data) || data[pos] != '<' || data[pos+1] != '<' {
		return pos + 1
	}
	i := pos + 2
	depth := 1
	for i < len(data) && depth > 0 {
		switch {
		case data[i] == '(':
			i = SkipStringLiteral(data, i)
		case data[i] == '<':
			if i+1 < len(data) && data[i+1] == '<' {
				depth++
				i += 2
			} else {
				i = SkipHexString(data, i)
			}
		case data[i] == '>' && i+1 < len(data) && data[i+1] == '>':
			depth--
			i += 2
		default:
			i++
		}
	}
	return i
}

// SkipArray skips a PDF array [...].
func SkipArray(data []byte, pos int) int {
	if pos >= len(data) || data[pos] != '[' {
		return pos + 1
	}
	i := pos + 1
	depth := 1
	for i < len(data) && depth > 0 {
		if data[i] == '(' {
			i = SkipStringLiteral(data, i)
			continue
		}
		if data[i] == '<' {
			if i+1 < len(data) && data[i+1] == '<' {
				i = SkipDictionary(data, i)
			} else {
				i = SkipHexString(data, i)
			}
			continue
		}
		switch data[i] {
		case '[':
			depth++
		case ']':
			depth--
		}
		i++
	}
	return i
}

// FindStreamStart finds the start of a stream in data, skipping strings/dicts.
// Returns index of "stream" keyword, or -1.
func FindStreamStart(data []byte) int {
	i := 0
	n := len(data)
	for i < n {
		if i+6 <= n && string(data[i:i+6]) == "stream" {
			if i+6 < n {
				b := data[i+6]
				if b == '\r' || b == '\n' {
					return i
				}
			}
		}

		switch {
		case data[i] == '(':
			i = SkipStringLiteral(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] != '<':
			i = SkipHexString(data, i)
		case data[i] == '<' && i+1 < n && data[i+1] == '<':
			i = SkipDictionary(data, i)
		case data[i] == '[':
			i = SkipArray(data, i)
		default:
			i++
		}
	}
	return -1
}

// ReplaceRefsInDictPart rewrites indirect references in pre-stream bytes
// while leaving literal (...) and hex <...> string contents untouched.
func ReplaceRefsInDictPart(pre []byte, replaceFunc func([]byte) []byte) []byte {
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	var out bytes.Buffer
	i := 0
	n := len(pre)
	segStart := 0
	flush := func(end int) {
		if end > segStart {
			out.Write(refRe.ReplaceAllFunc(pre[segStart:end], replaceFunc))
		}
	}
	for i < n {
		switch {
		case pre[i] == '(':
			flush(i)
			j := SkipStringLiteral(pre, i)
			if j > n {
				j = n
			}
			out.Write(pre[i:j])
			i, segStart = j, j
		case pre[i] == '<' && i+1 < n && pre[i+1] == '<':
			i += 2
		case pre[i] == '<':
			flush(i)
			j := SkipHexString(pre, i)
			if j > n {
				j = n
			}
			out.Write(pre[i:j])
			i, segStart = j, j
		default:
			i++
		}
	}
	flush(n)
	return out.Bytes()
}

// ReplaceRefsOutsideStreams rewrites indirect references only outside stream blocks.
func ReplaceRefsOutsideStreams(data []byte, offset int) []byte {
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	var out bytes.Buffer
	i := 0
	n := len(data)

	replaceFunc := func(b []byte) []byte {
		sm := refRe.FindSubmatch(b)
		if len(sm) < 3 {
			return b
		}
		on, err := strconv.Atoi(string(sm[1]))
		if err != nil {
			return b
		}
		gen := string(sm[2])
		return []byte(strconv.Itoa(offset+on) + " " + gen + " R")
	}

	for i < n {
		relStart := FindStreamStart(data[i:])
		if relStart == -1 {
			tail := data[i:]
			out.Write(ReplaceRefsInDictPart(tail, replaceFunc))
			break
		}

		streamStart := i + relStart

		pre := data[i:streamStart]
		out.Write(ReplaceRefsInDictPart(pre, replaceFunc))

		ptr := streamStart + 6
		if ptr < n && data[ptr] == '\r' {
			ptr++
		}
		if ptr < n && data[ptr] == '\n' {
			ptr++
		}

		endstreamIdx := -1
		searchPos := ptr
		for {
			idx := bytes.Index(data[searchPos:], []byte("endstream"))
			if idx == -1 {
				break
			}

			pos := searchPos + idx

			valid := false
			if pos > 0 {
				b := data[pos-1]
				if b == '\r' || b == '\n' {
					valid = true
				}
			}

			if valid {
				endstreamIdx = pos
				break
			}

			searchPos = pos + 9
		}

		if endstreamIdx == -1 {
			out.Write(data[streamStart:])
			break
		}

		endPos := endstreamIdx + 9
		out.Write(data[streamStart:endPos])
		i = endPos
	}

	return out.Bytes()
}

// HasSubstring checks if data contains substring.
func HasSubstring(data, sub []byte) bool {
	return bytes.Contains(data, sub)
}

// IsPageObject checks if the object body is a Page object.
func IsPageObject(body []byte) bool {
	return HasSubstring(body, []byte("/Type /Page")) ||
		HasSubstring(body, []byte("/Type/Page")) ||
		(HasSubstring(body, []byte("/MediaBox")) && !IsPagesTreeObject(body))
}

// IsPagesTreeObject checks if the object is a Pages tree node.
func IsPagesTreeObject(body []byte) bool {
	return HasSubstring(body, []byte("/Type /Pages")) ||
		HasSubstring(body, []byte("/Type/Pages"))
}

// IsWidgetAnnotation checks if the object is a Widget annotation.
func IsWidgetAnnotation(body []byte) bool {
	return HasSubstring(body, []byte("/Subtype /Widget")) ||
		HasSubstring(body, []byte("/Subtype/Widget"))
}

// IsFormField checks if the object has form field type.
func IsFormField(body []byte) bool {
	return HasSubstring(body, []byte("/FT /")) ||
		HasSubstring(body, []byte("/FT/"))
}

// IsXObjectForm checks if object is a Form XObject (appearance stream).
func IsXObjectForm(body []byte) bool {
	return HasSubstring(body, []byte("/Type /XObject")) &&
		HasSubstring(body, []byte("/Subtype /Form"))
}

// IsObjectStream checks if the object is an Object Stream (ObjStm).
func IsObjectStream(body []byte) bool {
	return HasSubstring(body, []byte("/Type /ObjStm")) ||
		HasSubstring(body, []byte("/Type/ObjStm"))
}

// IsWhitespace checks if byte is PDF whitespace.
func IsWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// ParseObjectStream extracts objects from a compressed object stream.
func ParseObjectStream(body []byte) map[int][]byte {
	result := make(map[int][]byte)

	nRe := regexp.MustCompile(`/N\s+(\d+)`)
	nMatch := nRe.FindSubmatch(body)
	if nMatch == nil {
		return result
	}
	numObjects, _ := strconv.Atoi(string(nMatch[1]))
	if numObjects <= 0 || numObjects > 50000 {
		return result
	}

	firstRe := regexp.MustCompile(`/First\s+(\d+)`)
	firstMatch := firstRe.FindSubmatch(body)
	if firstMatch == nil {
		return result
	}
	firstOffset, err := strconv.Atoi(string(firstMatch[1]))
	if err != nil || firstOffset < 0 {
		return result
	}

	streamData := ExtractAndDecompressStream(body)
	if streamData == nil || len(streamData) < firstOffset {
		return result
	}

	header := streamData[:firstOffset]
	objectData := streamData[firstOffset:]

	type objEntry struct {
		objNum int
		offset int
	}
	var entries []objEntry

	headerStr := string(bytes.TrimSpace(header))
	parts := regexp.MustCompile(`\s+`).Split(headerStr, -1)

	for i := 0; i+1 < len(parts); i += 2 {
		objNum, err1 := strconv.Atoi(parts[i])
		offset, err2 := strconv.Atoi(parts[i+1])
		if err1 == nil && err2 == nil {
			entries = append(entries, objEntry{objNum: objNum, offset: offset})
		}
	}

	for i, entry := range entries {
		start := entry.offset
		var end int
		if i+1 < len(entries) {
			end = entries[i+1].offset
		} else {
			end = len(objectData)
		}

		if start >= 0 && end <= len(objectData) && start < end {
			objBody := bytes.TrimSpace(objectData[start:end])
			result[entry.objNum] = objBody
		}
	}

	return result
}

// ExtractAndDecompressStream extracts and decompresses a stream from an object body.
func ExtractAndDecompressStream(body []byte) []byte {
	streamStart := FindStreamStart(body)
	if streamStart == -1 {
		return nil
	}

	ptr := streamStart + 6
	if ptr < len(body) && body[ptr] == '\r' {
		ptr++
	}
	if ptr < len(body) && body[ptr] == '\n' {
		ptr++
	}

	endstreamIdx := bytes.Index(body[ptr:], []byte("endstream"))
	if endstreamIdx == -1 {
		return nil
	}

	streamContent := body[ptr : ptr+endstreamIdx]

	for len(streamContent) > 0 {
		last := streamContent[len(streamContent)-1]
		if last == '\r' || last == '\n' {
			streamContent = streamContent[:len(streamContent)-1]
		} else {
			break
		}
	}

	if HasSubstring(body[:streamStart], []byte("/Filter /FlateDecode")) ||
		HasSubstring(body[:streamStart], []byte("/Filter/FlateDecode")) {
		return DecompressFlateResult(streamContent)
	}

	return streamContent
}

// BytesWithoutStreams returns data with raw stream contents blanked so
// keyword scans only see dictionary/trailer context, never stream bytes.
func BytesWithoutStreams(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	i := 0
	for i < len(data) {
		rel := FindStreamStart(out[i:])
		if rel == -1 {
			break
		}
		start := i + rel
		ptr := start + 6
		if ptr < len(out) && out[ptr] == '\r' {
			ptr++
		}
		if ptr < len(out) && out[ptr] == '\n' {
			ptr++
		}
		idx := bytes.Index(out[ptr:], []byte("endstream"))
		if idx == -1 {
			for j := ptr; j < len(out); j++ {
				out[j] = ' '
			}
			break
		}
		for j := ptr; j < ptr+idx; j++ {
			out[j] = ' '
		}
		i = ptr + idx + 9
	}
	return out
}

// ReadUint reads bytes as unsigned integer.
func ReadUint(b []byte) uint64 {
	var v uint64
	for _, c := range b {
		v = (v << 8) | uint64(byte(c))
	}
	return v
}

// ParseArrayInts parses array values from a PDF dictionary for key (e.g. "/W").
func ParseArrayInts(dict []byte, key string) []int {
	re := regexp.MustCompile(key + `\s*\[(.*?)\]`)
	m := re.FindSubmatch(dict)
	if m == nil {
		return nil
	}
	inner := bytes.TrimSpace(m[1])
	if len(inner) == 0 {
		return nil
	}
	parts := bytes.Fields(inner)
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		var v int
		if _, err := fmt.Sscanf(string(p), "%d", &v); err == nil {
			res = append(res, v)
		}
	}
	return res
}

// LocateStreamSegment returns the raw stream byte range inside an object
// body, honouring /Length when it points exactly at endstream and falling
// back to the last endstream token otherwise.
func LocateStreamSegment(obj []byte) (start, end int, ok bool) {
	streamIdx := bytes.Index(obj, []byte("stream"))
	if streamIdx < 0 {
		return 0, 0, false
	}
	start = streamIdx + len("stream")
	if start < len(obj) && obj[start] == '\r' {
		start++
	}
	if start < len(obj) && obj[start] == '\n' {
		start++
	}

	if l := inlineLength(obj); l > 0 && start+l <= len(obj) {
		endByLen := start + l
		k := endByLen
		for k < len(obj) && (obj[k] == '\r' || obj[k] == '\n' || obj[k] == ' ' || obj[k] == '\t') {
			k++
		}
		if k < len(obj) && bytes.HasPrefix(obj[k:], []byte("endstream")) {
			return start, endByLen, true
		}
	}

	endstreamIdx := bytes.LastIndex(obj, []byte("endstream"))
	if endstreamIdx < 0 || endstreamIdx < start {
		return 0, 0, false
	}
	end = endstreamIdx
	for end > start && (obj[end-1] == '\r' || obj[end-1] == '\n') {
		end--
	}
	if end <= start {
		return 0, 0, false
	}
	return start, end, true
}

func inlineLength(obj []byte) int {
	if regexp.MustCompile(`/Length\s+\d+\s+\d+\s+R`).Find(obj) != nil {
		return 0
	}
	re := regexp.MustCompile(`/Length\s+(\d+)`)
	m := re.FindSubmatch(obj)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(string(m[1]))
	if err != nil || n <= 0 {
		return 0
	}
	return n
}
