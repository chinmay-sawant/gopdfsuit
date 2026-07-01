// Package form provides functionality for parsing XFDF and filling PDF forms.
package form

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
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

func init() {
	var b strings.Builder
	b.WriteString("[")
	for i, w := range helveticaWidths {
		b.WriteString(strconv.Itoa(w))
		if i < len(helveticaWidths)-1 {
			b.WriteByte(' ')
		}
	}
	b.WriteByte(']')
	helveticaWidthsStr = b.String()
}

// Common PDF token literals reused across hot paths.
var (
	pdfSubtypeWidget      = []byte("/Subtype/Widget")
	pdfSubtypeWidgetSpace = []byte("/Subtype /Widget")
	pdfTokenT             = []byte("/T")
	pdfFTBtn              = []byte("/FT /Btn")
	pdfFTTx               = []byte("/FT /Tx")
	pdfFTTxAlt            = []byte("/FT/Tx")
	pdfObjStm             = []byte("/ObjStm")
	pdfTypeObjStm         = []byte("/Type/ObjStm")
	pdfEncrypt            = []byte("/Encrypt")
	pdfWBracket           = []byte("/W[")
	pdfIndex              = []byte("/Index")
	pdfStartxref          = []byte("startxref")
	pdfDictClose          = []byte(">>")
	pdfAP                 = []byte("/AP")
	pdfFieldsBracket      = []byte("/Fields[")
	pdfFieldsSpace        = []byte("/Fields [")
	pdfNeedAppearances    = []byte("/NeedAppearances")
)

// Package-scoped compiled regular expressions (PERF-1).
var (
	reValueToken        = regexp.MustCompile(`/V\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	reASToken           = regexp.MustCompile(`/AS\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	reRootRef           = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	reAcroFormRef       = regexp.MustCompile(`/AcroForm\s+(\d+)\s+(\d+)\s+R`)
	reParenString       = regexp.MustCompile(`\(([^)]{1,200})\)`)
	reHexString         = regexp.MustCompile(`<([0-9A-Fa-f\s]{2,400})>`)
	rePDFName           = regexp.MustCompile(`/([A-Za-z0-9_+-]{1,200})`)
	reTField            = regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	reKidsArray         = regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	reObjRef            = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	reSingleKidsRef     = regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	reVIndirectRef      = regexp.MustCompile(`/V\s*(\d+)\s+(\d+)\s+R`)
	reStreamBody        = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	reASWidget          = regexp.MustCompile(`/AS\s*(/(\w+)|\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	reAPDict            = regexp.MustCompile(`/AP\s*<<(.*?)>>`)
	reNSubDict          = regexp.MustCompile(`/N\s*<<(.*?)>>`)
	reAPKey             = regexp.MustCompile(`/([A-Za-z0-9_+-]+)\s*(?:/|stream|<<|\()`)
	reNName             = regexp.MustCompile(`/N\s*/([A-Za-z0-9_+-]+)`)
	reTrailer           = regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	reObjStream         = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	reObjStreamAlt      = regexp.MustCompile(`(?s)stream[\r\n]+(.*?)(?:[\r\n]+endstream|endstream)`)
	reObjAtOffset       = regexp.MustCompile(`(?s)^(\s*)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	reFirst             = regexp.MustCompile(`/First\s+(\d+)`)
	reFieldsArray       = regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
	reSingleFieldsRef   = regexp.MustCompile(`/Fields\s+(\d+)\s+(\d+)\s+R`)
	reWidgetDict        = regexp.MustCompile(`(?s)<<.*?/Subtype\s*/Widget.*?>>`)
	reWidgetName        = regexp.MustCompile(`/T\s*\((.*?)\)`)
	reAPNState          = regexp.MustCompile(`/AP\s*<<.*?/N\s*<<\s*/\s*([A-Za-z0-9_]+)\s*`)
	reRect              = regexp.MustCompile(`/Rect\s*\[\s*([^\]]+)\s*\]`)
	reQAlign            = regexp.MustCompile(`/Q\s*(\d)`)
	reDA                = regexp.MustCompile(`/DA\s*\((.*?)\)`)
	reTextFont          = regexp.MustCompile(`/([\w.-]+)\s+([\d.]+)\s+Tf`)
	reVGeneric          = regexp.MustCompile(`/V\s*\(?.*?\)?`)
	reVString           = regexp.MustCompile(`/V\s*\(.*?\)`)
	reAPStrip           = regexp.MustCompile(`\s*/AP\s*<<.*?>>`)
	reASState           = regexp.MustCompile(`/AS\s*/\w+`)
	reObjNum            = regexp.MustCompile(`(\d+)\s+0\s+obj`)
	reNeedAppearances   = regexp.MustCompile(`/NeedAppearances\s+(true|false)`)
	reAcroFormInline    = regexp.MustCompile(`(/AcroForm\s*<<)`)
	reRootObj           = regexp.MustCompile(`/Root\s+(\d+)\s+0\s+R`)
	reAcroFormIndirect  = regexp.MustCompile(`(?s)(/AcroForm\s*<<.*?)(>>)|(/AcroForm\s+\d+\s+\d+\s+R)`)
	reObjStmName        = regexp.MustCompile(`/T\s*(?:\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	reAPIndirectRef     = regexp.MustCompile(`/AP\s+\d+\s+\d+\s+R`)
	reLength            = regexp.MustCompile(`/Length\s+\d+`)
)

// XFDF structures for minimal parsing
type xfdfField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}

type xfdfRoot struct {
	XMLName xml.Name    `xml:"xfdf"`
	Fields  []xfdfField `xml:"fields>field"`
}

// ParseXFDF parses XFDF bytes and returns a map of field name -> value
func ParseXFDF(xfdfBytes []byte) (map[string]string, error) {
	var root xfdfRoot
	if err := xml.Unmarshal(xfdfBytes, &root); err != nil {
		return nil, err
	}
	m := make(map[string]string, len(root.Fields))
	for _, f := range root.Fields {
		name := strings.TrimSpace(f.Name)
		val := strings.TrimSpace(f.Value)
		m[name] = val
	}
	return m, nil
}

// Field represents a detected or targetable PDF form field.
type Field struct {
	Name  string
	Value string
	Type  string // V, AS, or detected type
}

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return bytes.Index(b, sub)
}

func objMapKey(objNum int, generation int) string {
	var b strings.Builder
	b.Grow(16)
	b.WriteString(strconv.Itoa(objNum))
	b.WriteByte(' ')
	b.WriteString(strconv.Itoa(generation))
	return b.String()
}

// splitFloatCoords splits PDF rect coordinate bytes without allocating via strings.Fields.
func splitFloatCoords(coords []byte) []string {
	trimmed := bytes.TrimSpace(coords)
	if len(trimmed) == 0 {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i <= len(trimmed); i++ {
		if i == len(trimmed) || unicode.IsSpace(rune(trimmed[i])) {
			if i > start {
				parts = append(parts, string(trimmed[start:i]))
			}
			start = i + 1
		}
	}
	return parts
}

func isTruthyButtonValue(val string) bool {
	return strings.EqualFold(val, "yes") || strings.EqualFold(val, "on")
}

// findEnclosingDict locates a PDF dictionary containing marker bytes.
func findEnclosingDict(pdf []byte, marker []byte, mustContain, mustNotContain []byte) (int, int, bool) {
	searchFrom := 0
	for {
		idx := bytes.Index(pdf[searchFrom:], marker)
		if idx < 0 {
			return 0, 0, false
		}
		abs := searchFrom + idx
		dictStart := bytes.LastIndex(pdf[:abs], []byte("<<"))
		if dictStart < 0 {
			searchFrom = abs + 1
			continue
		}
		relEnd := bytes.Index(pdf[abs:], pdfDictClose)
		if relEnd < 0 {
			searchFrom = abs + 1
			continue
		}
		dictEnd := abs + relEnd + len(pdfDictClose)
		slice := pdf[dictStart:dictEnd]
		if len(mustContain) > 0 && !bytes.Contains(slice, mustContain) {
			searchFrom = abs + 1
			continue
		}
		if len(mustNotContain) > 0 && bytes.Contains(slice, mustNotContain) {
			searchFrom = abs + 1
			continue
		}
		return dictStart, dictEnd, true
	}
}

// decodeHexString converts hex string to regular string
func decodeHexString(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	if len(s)%2 == 1 {
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return "<invalid hex>"
	}
	return string(b)
}

// tryZlibDecompress attempts to decompress zlib data
func tryZlibDecompress(b []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer func() {
		// Close error ignored: decompression already completed via io.Copy.
		_ = r.Close()
	}()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// tryFlateDecompress attempts to decompress raw flate data
func tryFlateDecompress(b []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(b))
	defer func() {
		// Close error ignored: decompression already completed via io.Copy.
		_ = r.Close()
	}()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// extractTokenGroups looks for /V or /AS tokens near a position
func extractTokenGroups(content []byte, pos int) (string, string) {
	limit := pos + 800
	if limit > len(content) {
		limit = len(content)
	}
	window := content[pos:limit]

	if m := reValueToken.FindSubmatch(window); m != nil {
		if len(m[2]) > 0 {
			return "V", string(m[2])
		}
		if len(m[3]) > 0 {
			return "V", decodeHexString(string(m[3]))
		}
		if len(m[4]) > 0 {
			return "V", string(m[4])
		}
	}
	if m := reASToken.FindSubmatch(window); m != nil {
		if len(m[2]) > 0 {
			return "AS", string(m[2])
		}
		if len(m[3]) > 0 {
			return "AS", decodeHexString(string(m[3]))
		}
		if len(m[4]) > 0 {
			return "AS", string(m[4])
		}
	}
	return "", ""
}

// parseArrayInts parses array values from PDF dictionary
func parseArrayInts(dict []byte, key string) []int {
	keyBytes := []byte(key)
	idx := bytes.Index(dict, keyBytes)
	if idx < 0 {
		return nil
	}
	rest := dict[idx+len(keyBytes):]
	bracketIdx := bytes.IndexByte(rest, '[')
	if bracketIdx < 0 {
		return nil
	}
	rest = rest[bracketIdx+1:]
	endIdx := bytes.IndexByte(rest, ']')
	if endIdx < 0 {
		return nil
	}
	inner := strings.TrimSpace(string(rest[:endIdx]))
	if inner == "" {
		return nil
	}
	parts := strings.Fields(inner)
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.Atoi(p); err == nil {
			res = append(res, v)
		}
	}
	return res
}

// readUint reads bytes as unsigned integer
func readUint(b []byte) uint64 {
	var v uint64
	for _, c := range b {
		v = (v << 8) | uint64(byte(c))
	}
	return v
}

// findRootRef looks for /Root n m R in the PDF bytes
func findRootRef(data []byte) (string, bool) {
	if m := reRootRef.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// getAcroFormRef finds /AcroForm n m R reference
func getAcroFormRef(body []byte, data []byte) (string, bool) {
	if m := reAcroFormRef.FindSubmatch(body); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	if m := reAcroFormRef.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// extractStringFromBytes looks for PDF literal representations
func extractStringFromBytes(b []byte) string {
	if m := reParenString.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	if m := reHexString.FindSubmatch(b); m != nil {
		return decodeHexString(string(m[1]))
	}
	if m := rePDFName.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	return ""
}

// traverseField resolves a field object and extracts field names and values
func traverseField(ref string, objMap map[string][]byte, parentPrefix string, out map[string]string) {
	body, ok := objMap[ref]
	if !ok {
		return
	}

	tv := ""
	name := ""
	if m := reTField.FindSubmatchIndex(body); m != nil {
		switch {
		case m[4] != -1 && m[5] != -1:
			name = string(body[m[4]:m[5]])
		case m[6] != -1 && m[7] != -1:
			name = decodeHexString(string(body[m[6]:m[7]]))
		case m[8] != -1 && m[9] != -1:
			name = string(body[m[8]:m[9]])
		}
		name = strings.TrimSpace(name)
		endPos := m[1]
		tType, val := extractTokenGroups(body, endPos)
		if tType != "" {
			tv = val
		} else {
			if rv := resolveValueRef(body, objMap); rv != "" {
				tv = rv
			}
			if tv == "" {
				if asn, ok := findWidgetAnnotationsForName(name, objMap); ok {
					tv = asn
				}
			}
		}
	}

	fullName := name
	if parentPrefix != "" && name != "" {
		fullName = parentPrefix + "." + name
	} else if parentPrefix != "" && name == "" {
		fullName = parentPrefix
	}

	if fullName != "" {
		if tv != "" {
			out[fullName] = tv
		} else if _, exists := out[fullName]; !exists {
			out[fullName] = ""
		}
	}

	if m := reKidsArray.FindSubmatch(body); m != nil {
		inner := m[1]
		for _, r := range reObjRef.FindAllSubmatch(inner, -1) {
			kidRef := string(r[1]) + " " + string(r[2])
			traverseField(kidRef, objMap, fullName, out)
		}
	}
	if m := reSingleKidsRef.FindSubmatch(body); m != nil {
		kidRef := string(m[1]) + " " + string(m[2])
		traverseField(kidRef, objMap, fullName, out)
	}
}

// resolveValueRef attempts to resolve /V value references
func resolveValueRef(body []byte, objMap map[string][]byte) string {
	var resolve func(b []byte, depth int) string
	resolve = func(b []byte, depth int) string {
		if depth > 6 {
			return ""
		}
		if tType, v := extractTokenGroups(b, 0); tType != "" && v != "" {
			return v
		}
		if s := extractStringFromBytes(b); s != "" {
			return s
		}
		if m := reVIndirectRef.FindSubmatch(b); m != nil {
			ref := string(m[1]) + " " + string(m[2])
			if rb, ok := objMap[ref]; ok {
				if sm := reStreamBody.FindSubmatch(rb); sm != nil {
					var dec []byte
					if d, err := tryZlibDecompress(sm[1]); err == nil {
						dec = d
					} else if d, err := tryFlateDecompress(sm[1]); err == nil {
						dec = d
					} else {
						dec = sm[1]
					}
					if s := extractStringFromBytes(dec); s != "" {
						return s
					}
				}
				if s := resolve(rb, depth+1); s != "" {
					return s
				}
				if s := extractStringFromBytes(rb); s != "" {
					return s
				}
				return "<resolved indirect>"
			}
		}
		return ""
	}
	return resolve(body, 0)
}

// findWidgetAnnotationsForName searches for widget annotations with the field name
func findWidgetAnnotationsForName(name string, objMap map[string][]byte) (string, bool) {
	needle := []byte("(" + name + ")")
	for k, body := range objMap {
		if bytesIndex(body, pdfSubtypeWidget) < 0 {
			continue
		}
		if bytesIndex(body, needle) < 0 && bytesIndex(body, pdfTokenT) < 0 {
			continue
		}
		if bytesIndex(body, needle) >= 0 {
			if m := reASWidget.FindSubmatch(body); m != nil {
				if len(m[2]) > 0 {
					return string(m[2]), true
				}
				if len(m[3]) > 0 {
					return string(m[3]), true
				}
				if len(m[4]) > 0 {
					return decodeHexString(string(m[4])), true
				}
			}
			if am := reAPDict.FindSubmatch(body); am != nil {
				if nm := reNSubDict.FindSubmatch(am[1]); nm != nil {
					if kr := reAPKey.FindSubmatch(nm[1]); kr != nil {
						return string(kr[1]), true
					}
				}
				if nn := reNName.FindSubmatch(am[1]); nn != nil {
					return string(nn[1]), true
				}
			}
			return k, true
		}
	}
	return "", false
}

// trailerHasEncrypt checks if trailer or any trailer 'Encrypt' appears
func trailerHasEncrypt(data []byte) bool {
	for _, m := range reTrailer.FindAllSubmatch(data, -1) {
		if bytesIndex(m[1], pdfEncrypt) >= 0 {
			return true
		}
	}
	// also check for /Encrypt elsewhere
	return bytesIndex(data, pdfEncrypt) >= 0
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap
func parseXRefStreams(data []byte, objMap map[string][]byte) {
	// find objects with streams that contain /W and /Index
	for _, m := range reObjStream.FindAllSubmatch(data, -1) {
		body := m[3]
		if bytesIndex(body, pdfWBracket) < 0 || bytesIndex(body, pdfIndex) < 0 {
			continue
		}
		// extract stream
		sm := reObjStreamAlt.FindSubmatch(body)
		if sm == nil {
			continue
		}
		streamBytes := sm[1]
		// decompress if needed
		var dec []byte
		if d, err := tryZlibDecompress(streamBytes); err == nil {
			dec = d
		} else if d, err := tryFlateDecompress(streamBytes); err == nil {
			dec = d
		} else {
			dec = streamBytes
		}

		// parse W and Index
		W := parseArrayInts(body, `/W`)
		if len(W) < 3 {
			continue
		}
		Index := parseArrayInts(body, `/Index`)
		if Index == nil {
			continue
		}

		// iterate index pairs
		w0, w1, w2 := W[0], W[1], W[2]
		total := w0 + w1 + w2
		for pos := 0; pos+total <= len(dec); pos += total {
			f1 := int(readUint(dec[pos : pos+w0]))
			f2 := int(readUint(dec[pos+w0 : pos+w0+w1]))
			f3 := int(readUint(dec[pos+w0+w1 : pos+total]))
			// type 1: f1==1 -> offset f3
			if f1 == 1 {
				off := f3
				if off > 0 && off < len(data) {
					// try to parse object at this offset
					tail := data[off:]
					if ro := reObjAtOffset.FindSubmatch(tail); ro != nil {
						onum := string(ro[2])
						ogen := string(ro[3])
						key := onum + " " + ogen
						objMap[key] = ro[4]
					}
				}
			}
			// type 2: object is in an object stream: f1==2 -> f2 is object stream number, f3 is index
			if f1 == 2 {
				objstm := f2
				index := f3
				// look for objstm content we earlier extracted
				key := objMapKey(objstm, 0)
				if stm, ok := objMap[key]; ok {
					// try to parse embedded objects from stm similarly to earlier logic
					_ = index
					_ = stm
				}
			}
		}
	}
}

// DetectFormFieldsAdvanced performs comprehensive field detection using the logic from fielddetect.go
//
//nolint:gocyclo
func DetectFormFieldsAdvanced(pdfBytes []byte) (map[string]string, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}

	// Check for encryption first
	if trailerHasEncrypt(pdfBytes) {
		// Fall back to naive detection for encrypted PDFs
		return detectFormFieldsNaive(pdfBytes)
	}

	// Build map of indirect objects
	objMatches := reObjStream.FindAllSubmatch(pdfBytes, -1)

	if len(objMatches) == 0 {
		// Fall back to naive scan
		return detectFormFieldsNaive(pdfBytes)
	}

	objMap := make(map[string][]byte, len(objMatches))
	for _, m := range objMatches {
		key := string(m[1]) + " " + string(m[2])
		body := m[3]

		// Handle ObjStm objects
		if bytesIndex(body, pdfObjStm) >= 0 || bytesIndex(body, pdfTypeObjStm) >= 0 {
			// find stream
			if sm := reObjStreamAlt.FindSubmatch(body); sm != nil {
				streamBytes := sm[1]
				// try decompress
				var dec []byte
				if d, err := tryZlibDecompress(streamBytes); err == nil {
					dec = d
				} else if d, err := tryFlateDecompress(streamBytes); err == nil {
					dec = d
				}
				if dec != nil {
					// find First value in dict
					first := 0
					if fm := reFirst.FindSubmatch(body); fm != nil {
						if n, err := strconv.Atoi(string(fm[1])); err == nil {
							first = n
						}
					}
					if first > 0 && first < len(dec) {
						// parse header portion up to first
						header := strings.TrimSpace(string(dec[:first]))
						parts := strings.Fields(header)
						// header should be pairs of (objnum offset)
						pairs := make([][2]int, 0, len(parts)/2)
						for i := 0; i+1 < len(parts); i += 2 {
							objnum, errNum := strconv.Atoi(parts[i])
							off, errOff := strconv.Atoi(parts[i+1])
							if errNum == nil && errOff == nil {
								pairs = append(pairs, [2]int{objnum, off})
							}
						}
						// objects content
						content := dec[first:]
						for pi := 0; pi < len(pairs); pi++ {
							objnum := pairs[pi][0]
							off := pairs[pi][1]
							var end int
							if pi+1 < len(pairs) {
								end = pairs[pi+1][1]
							} else {
								end = len(content)
							}
							if off < 0 || off >= len(content) || end <= off {
								continue
							}
							objBytes := content[off:end]
							// store under objnum 0 generation
							objMap[objMapKey(objnum, 0)] = objBytes
						}
						// also store the ObjStm object itself
						objMap[key] = body
						continue
					}
				}
			}
		}

		// Decompress streams if needed
		newBody := decompressStreams(body)
		objMap[key] = newBody
	}

	// Attempt to parse XRef streams to augment object map
	parseXRefStreams(pdfBytes, objMap)

	// Try structured approach first
	structured := make(map[string]string, len(objMap))
	if rootRef, ok := findRootRef(pdfBytes); ok {
		if rootBody, ok2 := objMap[rootRef]; ok2 {
			if acroRef, ok3 := getAcroFormRef(rootBody, pdfBytes); ok3 {
				if afBody, ok4 := objMap[acroRef]; ok4 {
					if fm := reFieldsArray.FindSubmatch(afBody); fm != nil {
						inner := fm[1]
						for _, r := range reObjRef.FindAllSubmatch(inner, -1) {
							fref := string(r[1]) + " " + string(r[2])
							traverseField(fref, objMap, "", structured)
						}
					} else {
						if sm := reSingleFieldsRef.FindSubmatch(afBody); sm != nil {
							fref := string(sm[1]) + " " + string(sm[2])
							traverseField(fref, objMap, "", structured)
						}
					}
				}
			}
		}
	}

	if len(structured) > 0 {
		return structured, nil
	}

	// Fall back to naive detection
	return detectFormFieldsNaive(pdfBytes)
}

// detectFormFieldsNaive performs simple field detection by scanning for /T tokens
func detectFormFieldsNaive(pdfBytes []byte) (map[string]string, error) {
	matches := reTField.FindAllSubmatchIndex(pdfBytes, -1)

	result := make(map[string]string, len(matches))
	seen := make(map[string]bool, len(matches))

	for _, mi := range matches {
		var name string
		switch {
		case mi[4] != -1 && mi[5] != -1:
			name = string(pdfBytes[mi[4]:mi[5]])
		case mi[6] != -1 && mi[7] != -1:
			name = decodeHexString(string(pdfBytes[mi[6]:mi[7]]))
		case mi[8] != -1 && mi[9] != -1:
			name = string(pdfBytes[mi[8]:mi[9]])
		default:
			continue
		}

		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		endPos := mi[1]
		tType, val := extractTokenGroups(pdfBytes, endPos)
		if tType != "" {
			result[name] = val
		} else {
			result[name] = ""
		}
	}

	return result, nil
}

// decompressStreams decompresses any compressed streams in the object body
func decompressStreams(body []byte) []byte {
	newBody := body

	for {
		found := false
		for _, sm := range reObjStreamAlt.FindAllSubmatchIndex(newBody, -1) {
			sStart := sm[2]
			sEnd := sm[3]
			if sStart < 0 || sEnd < 0 || sEnd <= sStart {
				continue
			}
			streamBytes := newBody[sStart:sEnd]
			var dec []byte
			if d, err := tryZlibDecompress(streamBytes); err == nil {
				dec = d
			} else if d, err := tryFlateDecompress(streamBytes); err == nil {
				dec = d
			}
			if dec != nil {
				var buf bytes.Buffer
				buf.Write(newBody[:sm[0]])
				buf.Write(dec)
				buf.Write(newBody[sm[1]:])
				newBody = buf.Bytes()
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	return newBody
}

// DetectFormFields now uses the enhanced detection logic
func DetectFormFields(pdfBytes []byte) ([]string, error) {
	fieldMap, err := DetectFormFieldsAdvanced(pdfBytes)
	if err != nil {
		return nil, err
	}

	var names []string
	for name := range fieldMap {
		names = append(names, name)
	}

	return names, nil
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

	mergedFields := make(map[string]string, len(detectedFields)+len(xfdfFields))
	for name, value := range detectedFields {
		mergedFields[name] = value
	}
	for name, value := range xfdfFields {
		mergedFields[name] = value
	}

	// Build a synthetic XFDF from merged fields so FillPDFWithXFDF can reuse logic
	genXFDF := buildXFDF(mergedFields)
	return FillPDFWithXFDF(pdfBytes, genXFDF)
}

// Helper to build minimal XFDF XML from field map
func buildXFDF(fields map[string]string) []byte {
	type xfdfField struct {
		XMLName xml.Name `xml:"field"`
		Name    string   `xml:"name,attr"`
		Value   string   `xml:"value"`
	}
	type xfdfRoot struct {
		XMLName xml.Name    `xml:"xfdf"`
		XMLNS   string      `xml:"xmlns,attr,omitempty"`
		Fields  []xfdfField `xml:"fields>field"`
	}
	root := xfdfRoot{XMLNS: "http://ns.adobe.com/xfdf/", Fields: make([]xfdfField, 0, len(fields))}
	for k, v := range fields {
		root.Fields = append(root.Fields, xfdfField{Name: k, Value: v})
	}
	out, err := xml.Marshal(root)
	if err != nil {
		return []byte(fmt.Sprintf(`<?xml version="1.0"?><xfdf xmlns="http://ns.adobe.com/xfdf/"><error>%s</error></xfdf>`, err.Error()))
	}
	return out
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
	widgetMatches := reWidgetDict.FindAllIndex(out, -1)

	discoveredWidgets := make(map[string][]job, len(fields))

	for _, match := range widgetMatches {
		dictStart, dictEnd := match[0], match[1]
		dictBytes := out[dictStart:dictEnd]
		nameMatch := reWidgetName.FindSubmatch(dictBytes)
		if nameMatch == nil {
			continue
		}
		fieldName := string(nameMatch[1])

		if _, ok := fields[fieldName]; !ok {
			continue // Skip if this field is not in our input data
		}

		currentVal := fields[fieldName]
		newJob := job{field: fieldName, val: currentVal, dictStart: dictStart, dictEnd: dictEnd}

		if bytes.Contains(dictBytes, pdfFTBtn) {
			if bytes.Contains(dictBytes, []byte("/Parent")) {
				newJob.fieldType = typeRadio
				if apMatch := reAPNState.FindSubmatch(dictBytes); apMatch != nil {
					newJob.radioExportValue = string(apMatch[1])
				}
			} else {
				newJob.fieldType = typeButton
			}
		} else if bytes.Contains(dictBytes, pdfFTTx) {
			newJob.fieldType = typeText
			rectMatch := reRect.FindSubmatch(dictBytes)
			if rectMatch == nil {
				continue
			}
			coords := splitFloatCoords(rectMatch[1])
			if len(coords) < 4 {
				continue
			}
			llx, err := strconv.ParseFloat(coords[0], 64)
			if err != nil {
				continue
			}
			lly, err := strconv.ParseFloat(coords[1], 64)
			if err != nil {
				continue
			}
			urx, err := strconv.ParseFloat(coords[2], 64)
			if err != nil {
				continue
			}
			ury, err := strconv.ParseFloat(coords[3], 64)
			if err != nil {
				continue
			}
			newJob.width, newJob.height = urx-llx, ury-lly

			if qMatch := reQAlign.FindSubmatch(dictBytes); qMatch != nil {
				if q, err := strconv.Atoi(string(qMatch[1])); err == nil {
					newJob.q = q
				}
			}
			newJob.fontSize, newJob.fontResourceName = 12.0, "Helv" // Default to Helv
			if daMatch := reDA.FindSubmatch(dictBytes); daMatch != nil {
				if tfMatch := reTextFont.FindStringSubmatch(string(daMatch[1])); len(tfMatch) > 2 {
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
	if bytes.Contains(out, pdfObjStm) {
		if filled, changed, err := fillXFDFInObjectStreams(out, fields); err != nil {
			return nil, err
		} else if changed {
			out = filled
			objStmChanged = true
		}
	}

	if len(allJobs) == 0 && !objStmChanged {
		return out, nil
	}

	if !objStmChanged {

		// --- PASS 2: MODIFICATION (WRITE-ONLY, IN REVERSE) ---
		sort.Slice(allJobs, func(i, j int) bool { return allJobs[i].dictStart > allJobs[j].dictStart })

		radioGroups := make(map[string]string, len(allJobs))
		for _, job := range allJobs {
			if job.fieldType == typeRadio {
				radioGroups[job.field] = job.val
			}
		}
		for fieldName, value := range radioGroups {
			marker := []byte("/T (" + fieldName + ")")
			dictStart, dictEnd, ok := findEnclosingDict(out, marker, []byte("/Kids"), pdfSubtypeWidgetSpace)
			if !ok {
				continue
			}
			dictBytes := out[dictStart:dictEnd]
			var newVBuf strings.Builder
			newVBuf.WriteString("/V /")
			newVBuf.WriteString(value)
			newV := []byte(newVBuf.String())
			var newDictBytes []byte
			if reVGeneric.Match(dictBytes) {
				newDictBytes = reVGeneric.ReplaceAll(dictBytes, newV)
			} else {
				newDictBytes = bytes.Replace(dictBytes, pdfDictClose, append(append([]byte(" "), newV...), pdfDictClose...), 1)
			}
			out = append(out[:dictStart], append(newDictBytes, out[dictEnd:]...)...)
		}

		for _, job := range allJobs {
			dictBytes := out[job.dictStart:job.dictEnd]
			var newDictBytes []byte
			switch job.fieldType {
			case typeText:
				esc := escapePDFString(job.val)
				var newVBuf strings.Builder
				newVBuf.WriteString("/V (")
				newVBuf.WriteString(esc)
				newVBuf.WriteByte(')')
				newV := []byte(newVBuf.String())
				if reVString.Match(dictBytes) {
					newDictBytes = reVString.ReplaceAll(dictBytes, newV)
				} else {
					newDictBytes = bytes.Replace(dictBytes, pdfDictClose, append(append([]byte(" "), newV...), pdfDictClose...), 1)
				}
				newDictBytes = reAPStrip.ReplaceAll(newDictBytes, []byte(" "))
			case typeButton, typeRadio:
				newState := "/Off"
				if job.fieldType == typeButton && isTruthyButtonValue(job.val) {
					if apMatch := reAPNState.FindSubmatch(dictBytes); apMatch != nil {
						newState = "/" + string(apMatch[1])
					} else {
						newState = "/Yes"
					}
				} else if job.fieldType == typeRadio && job.radioExportValue == job.val {
					newState = "/" + job.radioExportValue
				}
				newAS := []byte("/AS " + newState)
				if reASState.Match(dictBytes) {
					newDictBytes = reASState.ReplaceAll(dictBytes, newAS)
				} else {
					newDictBytes = bytes.Replace(dictBytes, pdfDictClose, append(append([]byte(" "), newAS...), pdfDictClose...), 1)
				}
			}
			if newDictBytes != nil {
				out = append(out[:job.dictStart], append(newDictBytes, out[job.dictEnd:]...)...)
			}
		}
	}

	// --- PASS 3: NEW OBJECT GENERATION ---
	allObjMatches := reObjNum.FindAllSubmatchIndex(out, -1)
	highest := 0
	for _, m := range allObjMatches {
		if n, err := strconv.Atoi(string(out[m[2]:m[3]])); err == nil && n > highest {
			highest = n
		}
	}
	nextObj := highest + 1
	if sx := bytes.LastIndex(out, pdfStartxref); sx >= 0 {
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
		marker := []byte("/T (" + job.field + ")")
		_, dictEnd, ok := findEnclosingDict(out, marker, pdfSubtypeWidget, nil)
		if !ok {
			continue
		}
		var apRefBuf strings.Builder
		apRefBuf.WriteString(" /AP<</N ")
		apRefBuf.WriteString(strconv.Itoa(job.apObjNum))
		apRefBuf.WriteString(" 0 R>>")
		apRef := []byte(apRefBuf.String())
		out = append(out[:dictEnd-2], append(apRef, out[dictEnd-2:]...)...)
	}

	if reNeedAppearances.Match(out) {
		out = reNeedAppearances.ReplaceAll(out, []byte("/NeedAppearances false"))
	} else {
		if loc := reAcroFormInline.FindIndex(out); loc != nil {
			insertPos := loc[1]
			insertContent := []byte(" /NeedAppearances false ")
			out = append(append(append(out[:insertPos:insertPos], insertContent...), out[insertPos:]...))
		}
	}

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
		var streamBuf strings.Builder
		streamBuf.WriteString("q\nBT\n/F1 ")
		streamBuf.WriteString(strconv.FormatFloat(job.fontSize, 'f', 2, 64))
		streamBuf.WriteString(" Tf\n0 g\n")
		streamBuf.WriteString(strconv.FormatFloat(tx, 'f', 2, 64))
		streamBuf.WriteByte(' ')
		streamBuf.WriteString(strconv.FormatFloat(y, 'f', 2, 64))
		streamBuf.WriteString(" Td\n(")
		streamBuf.WriteString(streamText)
		streamBuf.WriteString(") Tj\nET\nQ")
		streamBody := streamBuf.String()

		var fontDescBuf strings.Builder
		fontDescBuf.WriteString("\n")
		fontDescBuf.WriteString(strconv.Itoa(job.fontDescObjNum))
		fontDescBuf.WriteString(" 0 obj\n<</Type/FontDescriptor/FontName/")
		fontDescBuf.WriteString(job.fontResourceName)
		fontDescBuf.WriteString("/Flags 32/FontBBox[-558 -225 1000 931]/ItalicAngle 0/Ascent 905/Descent -212/CapHeight 905/StemV 88>>\nendobj\n")
		out = append(out, fontDescBuf.String()...)

		var fontObjBuf strings.Builder
		fontObjBuf.WriteString("\n")
		fontObjBuf.WriteString(strconv.Itoa(job.fontObjNum))
		fontObjBuf.WriteString(" 0 obj\n<</Type/Font/Subtype/Type1/BaseFont/")
		fontObjBuf.WriteString(job.fontResourceName)
		fontObjBuf.WriteString("/Encoding/WinAnsiEncoding/FirstChar 32/LastChar 255/Widths ")
		fontObjBuf.WriteString(helveticaWidthsStr)
		fontObjBuf.WriteString("/FontDescriptor ")
		fontObjBuf.WriteString(strconv.Itoa(job.fontDescObjNum))
		fontObjBuf.WriteString(" 0 R>>\nendobj\n")
		out = append(out, fontObjBuf.String()...)

		var compBuf bytes.Buffer
		zw, err := zlib.NewWriterLevel(&compBuf, zlib.BestCompression)
		if err != nil {
			zw = zlib.NewWriter(&compBuf)
		}
		if _, err := zw.Write([]byte(streamBody)); err != nil {
			return nil, fmt.Errorf("compression write failed: %w", err)
		}
		if err := zw.Close(); err != nil {
			return nil, fmt.Errorf("compression close failed: %w", err)
		}
		comp := compBuf.Bytes()
		var apObjBuf strings.Builder
		apObjBuf.WriteString("\n")
		apObjBuf.WriteString(strconv.Itoa(job.apObjNum))
		apObjBuf.WriteString(" 0 obj\n<</Type/XObject/Subtype/Form/FormType 1/BBox[0 0 ")
		apObjBuf.WriteString(strconv.FormatFloat(job.width, 'f', 2, 64))
		apObjBuf.WriteByte(' ')
		apObjBuf.WriteString(strconv.FormatFloat(job.height, 'f', 2, 64))
		apObjBuf.WriteString("/Resources<</Font<</F1 ")
		apObjBuf.WriteString(strconv.Itoa(job.fontObjNum))
		apObjBuf.WriteString(" 0 R>>/ProcSet[/PDF/Text]>>/Filter/FlateDecode/Length ")
		apObjBuf.WriteString(strconv.Itoa(len(comp)))
		apObjBuf.WriteString(">>\nstream\n")
		apObjBuf.Write(comp)
		apObjBuf.WriteString("\nendstream\nendobj\n")
		out = append(out, apObjBuf.String()...)
	}

	objMatches := reObjNum.FindAllSubmatchIndex(out, -1)
	offsets := make(map[int]int, len(objMatches))
	maxObj := 0
	for _, m := range objMatches {
		num, err := strconv.Atoi(string(out[m[2]:m[3]]))
		if err != nil {
			continue
		}
		offsets[num] = m[0]
		if num > maxObj {
			maxObj = num
		}
	}
	xrefStart := len(out)
	var xrefBuf strings.Builder
	xrefBuf.WriteString("xref\n0 ")
	xrefBuf.WriteString(strconv.Itoa(maxObj + 1))
	xrefBuf.WriteString("\n0000000000 65535 f \r\n")
	var xrefLine [32]byte
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			line := strconv.AppendInt(xrefLine[:0], int64(off), 10)
			padding := 10 - len(line)
			if padding > 0 {
				padded := make([]byte, 10)
				copy(padded[padding:], line)
				for j := 0; j < padding; j++ {
					padded[j] = '0'
				}
				line = padded
			}
			xrefBuf.Write(line)
			xrefBuf.WriteString(" 00000 n \r\n")
		} else {
			xrefBuf.WriteString("0000000000 65535 f \r\n")
		}
	}
	root := 1
	if rm := reRootObj.FindSubmatch(pdfBytes); len(rm) > 1 {
		if r, err := strconv.Atoi(string(rm[1])); err == nil {
			root = r
		}
	}
	var trailerBuf strings.Builder
	trailerBuf.WriteString("trailer\n<</Size ")
	trailerBuf.WriteString(strconv.Itoa(maxObj + 1))
	trailerBuf.WriteString("/Root ")
	trailerBuf.WriteString(strconv.Itoa(root))
	trailerBuf.WriteString(" 0 R>>\nstartxref\n")
	trailerBuf.WriteString(strconv.Itoa(xrefStart))
	trailerBuf.WriteString("\n%%EOF\n")
	out = append(out, xrefBuf.String()...)
	out = append(out, trailerBuf.String()...)
	// --- PASS 3: GLOBAL NEED APPEARANCES ---
	// If fields were modified or APs stripped, force the PDF viewer to recreate appearances on open.
	acroMatch := reAcroFormIndirect.FindSubmatch(out)
	if acroMatch != nil {
		if acroMatch[1] != nil {
			// Inline dictionary case
			dictPart := acroMatch[1]
			if reNeedAppearances.Match(dictPart) {
				newDict := reNeedAppearances.ReplaceAll(dictPart, []byte("/NeedAppearances true"))
				out = bytes.Replace(out, dictPart, newDict, 1)
			} else {
				// Inject it
				newDict := make([]byte, len(dictPart)+len(" /NeedAppearances true "))
				copy(newDict, dictPart)
				copy(newDict[len(dictPart):], " /NeedAppearances true ")
				out = bytes.Replace(out, dictPart, newDict, 1)
			}
		} else if acroMatch[3] != nil {
			// Indirect reference case
			refFull := string(acroMatch[3])
			if rm := reObjRef.FindStringSubmatch(refFull); rm != nil {
				objPattern := `(?s)\b` + rm[1] + `\s+` + rm[2] + `\s+obj(.*?)endobj`
				if objM := regexp.MustCompile(objPattern).FindSubmatch(out); objM != nil {
					objBody := objM[1]
					if reNeedAppearances.Match(objBody) {
						newBody := reNeedAppearances.ReplaceAll(objBody, []byte("/NeedAppearances true"))
						out = bytes.Replace(out, objBody, newBody, 1)
					} else {
						// Inject before the ending >>
						insertPos := bytes.LastIndex(objBody, pdfDictClose)
						if insertPos >= 0 {
							var newBody bytes.Buffer
							newBody.Write(objBody[:insertPos])
							newBody.WriteString(" /NeedAppearances true ")
							newBody.Write(objBody[insertPos:])
							out = bytes.Replace(out, objBody, newBody.Bytes(), 1)
						}
					}
				}
			}
		}
	}

	return out, nil
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

	matches := reObjStream.FindAllSubmatchIndex(out, -1)
	if len(matches) == 0 {
		return out, false, nil
	}

	changedAny := false
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		bodyStart, bodyEnd := m[6], m[7]
		body := out[bodyStart:bodyEnd]
		if !bytes.Contains(body, pdfObjStm) {
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
	sm := reObjStreamAlt.FindSubmatchIndex(body)
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

	kidsToRemoveAP := make(map[int]bool, len(members))
	for i := range members {
		start := members[i].offset
		end := len(content)
		if i+1 < len(members) {
			end = members[i+1].offset
		}
		objContent := bytes.TrimSpace(content[start:end])

		if nameMatch := reObjStmName.FindSubmatch(objContent); nameMatch != nil {
			var fieldName string
			if len(nameMatch[1]) > 0 {
				fieldName = string(nameMatch[1])
			} else if len(nameMatch[2]) > 0 {
				fieldName = decodeHexString(string(nameMatch[2]))
			}
			fieldName = strings.TrimSpace(fieldName)

			if _, ok := fields[fieldName]; ok {
				if bytes.Contains(objContent, pdfFTTx) || bytes.Contains(objContent, pdfFTTxAlt) {
					kidsToRemoveAP[members[i].objNum] = true
					if m := reKidsArray.FindSubmatch(objContent); m != nil {
						for _, r := range reObjRef.FindAllSubmatch(m[1], -1) {
							if kidNum, err := strconv.Atoi(string(r[1])); err == nil {
								kidsToRemoveAP[kidNum] = true
							}
						}
					}
					if m := reSingleKidsRef.FindSubmatch(objContent); m != nil {
						if kidNum, err := strconv.Atoi(string(m[1])); err == nil {
							kidsToRemoveAP[kidNum] = true
						}
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
			if reAPIndirectRef.Match(updated) {
				updated = reAPIndirectRef.ReplaceAll(updated, []byte(" "))
				changed = true
			}
			// Manual removal of /AP dictionary to handle nested <<>>
			apIdx := bytes.Index(updated, pdfAP)
			if apIdx >= 0 {
				afterAP := updated[apIdx+3:]
				trimmedAfter := bytes.TrimSpace(afterAP)
				if bytes.HasPrefix(trimmedAfter, []byte("<<")) {
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
		if bytes.Contains(updated, pdfFieldsBracket) || bytes.Contains(updated, pdfFieldsSpace) {
			// Basic AcroForm object identification
			if !bytes.Contains(updated, pdfNeedAppearances) {
				insertPos := bytes.LastIndex(updated, pdfDictClose)
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

	var headerBuilder strings.Builder
	var contentBuilder strings.Builder
	currentOffset := 0
	for i, member := range members {
		headerBuilder.WriteString(strconv.Itoa(member.objNum))
		headerBuilder.WriteByte(' ')
		headerBuilder.WriteString(strconv.Itoa(currentOffset))
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

	newHeader := headerBuilder.String()
	newFirst := len(newHeader) + 1
	newDecoded := []byte(newHeader + " " + contentBuilder.String())

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
	var firstBuf strings.Builder
	firstBuf.WriteString("/First ")
	firstBuf.WriteString(strconv.Itoa(newFirst))
	newDict := reFirst.ReplaceAll(dictPart, []byte(firstBuf.String()))
	var lengthBuf strings.Builder
	lengthBuf.WriteString("/Length ")
	lengthBuf.WriteString(strconv.Itoa(len(compressed)))
	if reLength.Match(newDict) {
		newDict = reLength.ReplaceAll(newDict, []byte(lengthBuf.String()))
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
	nameMatch := reObjStmName.FindSubmatch(objContent)
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

	if bytes.Contains(updated, pdfFTTx) || bytes.Contains(updated, pdfFTTxAlt) {
		replacement := []byte("/V (" + escapePDFString(value) + ")")
		newUpdated, changed := replaceOrInsertPDFEntry(updated, `/V\s*\((?:\\.|[^\\)])*\)|/V\s*/[^\s/>]+|/V\s*<[0-9A-Fa-f\s]+>`, replacement)
		return newUpdated, changed
	}

	if bytes.Contains(updated, pdfFTBtn) || bytes.Contains(updated, []byte("/FT/Btn")) {
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

	insertPos := bytes.LastIndex(dict, pdfDictClose)
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

// escapePDFString escapes characters as required for PDF literal strings.
func escapePDFString(s string) string {
	// Fast path: most text has no special characters
	if !strings.ContainsAny(s, `()\`) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s) + 4)
	for _, r := range s {
		switch r {
		case '(', ')', '\\':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
