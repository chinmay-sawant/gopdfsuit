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
)

var (
	reAS              = regexp.MustCompile(`/AS\s*(/(\w+)|\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	reAP              = regexp.MustCompile(`/AP\s*<<(.*?)>>`)
	reN               = regexp.MustCompile(`/N\s*<<(.*?)>>`)
	reKey             = regexp.MustCompile(`/([A-Za-z0-9_+-]+)\s*(?:/|stream|<<|\()`)
	reNName           = regexp.MustCompile(`/N\s*/([A-Za-z0-9_+-]+)`)
	reStream          = regexp.MustCompile(`(?s)stream[\r\n]+(.*?)(?:[\r\n]+endstream|endstream)`)
	reObj             = regexp.MustCompile(`(?s)^(\s*)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
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
	reTShort          = regexp.MustCompile(`/T\s*(?:\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	reTWidget         = regexp.MustCompile(`/T\s*\((.*?)\)`)
	reKids            = regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	reRef             = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	reSingleKids      = regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	reVRef            = regexp.MustCompile(`/V\s*(\d+)\s+(\d+)\s+R`)
	reStreamAlt       = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	reTrailer         = regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	reObjStream       = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	reFields          = regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
	reFieldsSingle    = regexp.MustCompile(`/Fields\s+(\d+)\s+(\d+)\s+R`)
	reWidget          = regexp.MustCompile(`(?s)<<.*?/Subtype\s*/Widget.*?>>`)
	reRect            = regexp.MustCompile(`/Rect\s*\[\s*([^\]]+)\s*\]`)
	reQ               = regexp.MustCompile(`/Q\s*(\d)`)
	reDA              = regexp.MustCompile(`/DA\s*\((.*?)\)`)
	reTf              = regexp.MustCompile(`/([\w.-]+)\s+([\d.]+)\s+Tf`)
	reVBroad          = regexp.MustCompile(`/V\s*\(?.*?\)?`)
	reVParen          = regexp.MustCompile(`/V\s*\((?:\\.|[^\\)])*\)`)
	reBtnOnState      = regexp.MustCompile(`/AP\s*<<.*?/N\s*<<[^>]*?/Yes`)
	reObj0            = regexp.MustCompile(`(\d+)\s+0\s+obj`)
	reNeedAppearances = regexp.MustCompile(`/NeedAppearances\s+(true|false)`)
	reAcroForm        = regexp.MustCompile(`(/AcroForm\s*<<)`)
	reRoot0           = regexp.MustCompile(`/Root\s+(\d+)\s+0\s+R`)
	reAcroFormBoth    = regexp.MustCompile(`(?s)(/AcroForm\s*<<.*?)(>>)|(/AcroForm\s+\d+\s+\d+\s+R)`)
	reAPRef           = regexp.MustCompile(`/AP\s+\d+\s+\d+\s+R`)
	reLength          = regexp.MustCompile(`/Length\s+\d+`)
	reAcroFormRef     = regexp.MustCompile(`/AcroForm\s+(\d+)\s+(\d+)\s+R`)
	reRootRef         = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)

	bytesSubtypeWidget      = []byte("/Subtype/Widget")
	bytesSubtypeSpaceWidget = []byte("/Subtype /Widget")
	bytesT                  = []byte("/T")
	bytesW                  = []byte("/W[")
	bytesIdx                = []byte("/Index")
	bytesEncrypt            = []byte("/Encrypt")
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
	m := make(map[string]string)
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

// findBalancedDictBounds returns the [start, end) slice of a PDF dictionary opened at openIdx (<<).
func findBalancedDictBounds(data []byte, openIdx int) (start, end int, ok bool) {
	if openIdx < 0 || openIdx+1 >= len(data) || data[openIdx] != '<' || data[openIdx+1] != '<' {
		return 0, 0, false
	}
	depth := 0
	for i := openIdx; i < len(data)-1; i++ {
		if data[i] == '<' && data[i+1] == '<' {
			depth++
			i++
			continue
		}
		if data[i] == '>' && data[i+1] == '>' {
			depth--
			i++
			if depth == 0 {
				return openIdx, i + 1, true
			}
		}
	}
	return 0, 0, false
}

// findWidgetDictBounds locates the widget annotation dictionary containing /T (fieldName).
func findWidgetDictBounds(data []byte, fieldName string) (start, end int, ok bool) {
	tKey := []byte("/T (" + fieldName + ")")
	tIdx := bytes.Index(data, tKey)
	if tIdx < 0 {
		return 0, 0, false
	}
	depth := 0
	openIdx := -1
	for i := tIdx; i >= 1; i-- {
		if data[i-1] == '<' && data[i] == '<' {
			if depth == 0 {
				openIdx = i - 1
				break
			}
			depth--
			i--
			continue
		}
		if data[i-1] == '>' && data[i] == '>' {
			depth++
			i--
		}
	}
	if openIdx < 0 {
		return 0, 0, false
	}
	start, end, ok = findBalancedDictBounds(data, openIdx)
	if !ok {
		return 0, 0, false
	}
	dict := data[start:end]
	if !bytes.Contains(dict, bytesSubtypeSpaceWidget) && !bytes.Contains(dict, bytesSubtypeWidget) {
		return 0, 0, false
	}
	return start, end, true
}

func buttonOnState(dictBytes []byte) string {
	if reBtnOnState.Match(dictBytes) {
		return "/Yes"
	}
	return "/Yes"
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

	if m := reValue.FindSubmatch(window); m != nil {
		if len(m[1]) > 2 && m[1][0] == '(' && m[1][len(m[1])-1] == ')' {
			return "V", unescapePDFString(string(m[1][1 : len(m[1])-1]))
		}
		if len(m[2]) > 0 {
			return "V", decodeHexString(string(m[2]))
		}
		if len(m[3]) > 0 {
			return "V", string(m[3])
		}
	}
	if m := reAsToken.FindSubmatch(window); m != nil {
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
	re := regexp.MustCompile(key + `\s*\[(.*?)\]`)
	if m := re.FindSubmatch(dict); m != nil {
		inner := strings.TrimSpace(string(m[1]))
		if inner == "" {
			return nil
		}
		parts := strings.Fields(inner)
		res := make([]int, 0, len(parts))
		for _, p := range parts {
			var v int
			if _, err := fmt.Sscanf(p, "%d", &v); err == nil {
				res = append(res, v)
			}
		}
		return res
	}
	return nil
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
	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	if m := rootRe.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// getAcroFormRef finds /AcroForm n m R reference
func getAcroFormRef(body []byte, data []byte) (string, bool) {
	afRe := regexp.MustCompile(`/AcroForm\s+(\d+)\s+(\d+)\s+R`)
	if m := afRe.FindSubmatch(body); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	if m := afRe.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// extractStringFromBytes looks for PDF literal representations
func extractStringFromBytes(b []byte) string {
	if m := reParen.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	if m := reHexString.FindSubmatch(b); m != nil {
		return decodeHexString(string(m[1]))
	}
	if m := reNameStr.FindSubmatch(b); m != nil {
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
	if m := reTFull.FindSubmatchIndex(body); m != nil {
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

	if m := reKids.FindSubmatch(body); m != nil {
		inner := m[1]
		for _, r := range reRef.FindAllSubmatch(inner, -1) {
			kidRef := string(r[1]) + " " + string(r[2])
			traverseField(kidRef, objMap, fullName, out)
		}
	}
	if m := reSingleKids.FindSubmatch(body); m != nil {
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
		if m := reVRef.FindSubmatch(b); m != nil {
			ref := string(m[1]) + " " + string(m[2])
			if rb, ok := objMap[ref]; ok {
				if sm := reStreamAlt.FindSubmatch(rb); sm != nil {
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
		if bytesIndex(body, bytesSubtypeWidget) < 0 {
			continue
		}
		if bytesIndex(body, needle) < 0 && bytesIndex(body, bytesT) < 0 {
			continue
		}
		if bytesIndex(body, needle) >= 0 {
			if m := reAS.FindSubmatch(body); m != nil {
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
		if am := reAP.FindSubmatch(body); am != nil {
				if nm := reN.FindSubmatch(am[1]); nm != nil {
					if kr := reKey.FindSubmatch(nm[1]); kr != nil {
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
		if bytesIndex(m[1], bytesEncrypt) >= 0 {
			return true
		}
	}
	// also check for /Encrypt elsewhere
	return bytesIndex(data, bytesEncrypt) >= 0
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap
func parseXRefStreams(data []byte, objMap map[string][]byte) {
	// find objects with streams that contain /W and /Index
	for _, m := range reObjStream.FindAllSubmatch(data, -1) {
		body := m[3]
		if bytesIndex(body, bytesW) < 0 || bytesIndex(body, bytesIdx) < 0 {
			continue
		}
		// extract stream
		sm := reStream.FindSubmatch(body)
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
					if ro := reObj.FindSubmatch(tail); ro != nil {
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
				var buf [30]byte
				b := strconv.AppendInt(buf[:0], int64(objstm), 10)
				b = append(b, ' ', '0')
				key := string(b)
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

	objMap := make(map[string][]byte)
	for _, m := range objMatches {
		key := string(m[1]) + " " + string(m[2])
		body := m[3]

		// Handle ObjStm objects
		if bytesIndex(body, []byte("/ObjStm")) >= 0 || bytesIndex(body, []byte("/Type/ObjStm")) >= 0 {
			// find stream
			if sm := reStream.FindSubmatch(body); sm != nil {
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
						pairs := [][]int{}
						for i := 0; i+1 < len(parts); i += 2 {
							var objnum, off int
							if _, err := fmt.Sscanf(parts[i], "%d", &objnum); err == nil {
								if _, err2 := fmt.Sscanf(parts[i+1], "%d", &off); err2 == nil {
									pairs = append(pairs, []int{objnum, off})
								}
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
							var objKeyBuf [30]byte
							b := strconv.AppendInt(objKeyBuf[:0], int64(objnum), 10)
							b = append(b, ' ', '0')
							objKey := string(b)
							objMap[objKey] = objBytes
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
	structured := make(map[string]string)
	if rootRef, ok := findRootRef(pdfBytes); ok {
		if rootBody, ok2 := objMap[rootRef]; ok2 {
			if acroRef, ok3 := getAcroFormRef(rootBody, pdfBytes); ok3 {
				if afBody, ok4 := objMap[acroRef]; ok4 {
					fieldsRe := regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
					if fm := fieldsRe.FindSubmatch(afBody); fm != nil {
						inner := fm[1]
						refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
						for _, r := range refRe.FindAllSubmatch(inner, -1) {
							fref := string(r[1]) + " " + string(r[2])
							traverseField(fref, objMap, "", structured)
						}
					} else {
						singleFields := regexp.MustCompile(`/Fields\s+(\d+)\s+(\d+)\s+R`)
						if sm := singleFields.FindSubmatch(afBody); sm != nil {
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
	matches := reTFull.FindAllSubmatchIndex(pdfBytes, -1)

	result := make(map[string]string)
	seen := make(map[string]bool)

	for _, mi := range matches {
		var name string
		switch {
		case mi[4] != -1 && mi[5] != -1:
			name = unescapePDFString(string(pdfBytes[mi[4]:mi[5]]))
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
	streamRe := regexp.MustCompile(`(?s)stream[\r\n]+(.*?)(?:[\r\n]+endstream|endstream)`)
	newBody := body

	for {
		found := false
		for _, sm := range streamRe.FindAllSubmatchIndex(newBody, -1) {
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

	mergedFields := make(map[string]string)
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
	out, _ := xml.Marshal(root)
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

	if len(allJobs) == 0 && !objStmChanged {
		return out, nil
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
		newDict := append(dictBytes[:len(dictBytes)-2], append(apRef, dictBytes[len(dictBytes)-2:]...)...)
		out = append(out[:dictStart], append(newDict, out[dictEnd:]...)...)
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
		buf.WriteString(buildWidthsStr())
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
	// --- PASS 3: GLOBAL NEED APPEARANCES ---
	// If fields were modified or APs stripped, force the PDF viewer to recreate appearances on open.
	acroMatch := reAcroFormBoth.FindSubmatch(out)
	if acroMatch != nil {
		if acroMatch[1] != nil {
			// Inline dictionary case
			dictPart := acroMatch[1]
			if reNeedAppearances.Match(dictPart) {
				newDict := reNeedAppearances.ReplaceAll(dictPart, bytesNeedAppTrue)
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
			if rm := reRef.FindStringSubmatch(refFull); rm != nil {
				objRe := regexp.MustCompile(`(?s)\b` + rm[1] + `\s+` + rm[2] + `\s+obj(.*?)endobj`)
				if objM := objRe.FindSubmatch(out); objM != nil {
					objBody := objM[1]
					if reNeedAppearances.Match(objBody) {
						newBody := reNeedAppearances.ReplaceAll(objBody, bytesNeedAppTrue)
						out = bytes.Replace(out, objBody, newBody, 1)
					} else {
						// Inject before the ending >>
						insertPos := bytes.LastIndex(objBody, bytesGtGt)
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

// unescapePDFString reverses escapePDFString for literal PDF strings.
func unescapePDFString(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			sb.WriteByte(s[i])
			continue
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
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
