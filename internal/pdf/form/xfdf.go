// Package form provides functionality for parsing XFDF and filling PDF forms.
package form

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"maps"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/byteconv"
)

// Package-level PDF markers (avoid per-call []byte allocations).
var (
	markerEncrypt       = []byte(`/Encrypt`)
	markerWBracket      = []byte(`/W[`)
	markerIndex         = []byte(`/Index`)
	markerSubtypeWidget = []byte(`/Subtype/Widget`)
	markerT             = []byte(`/T`)
	markerObjStm        = []byte("/ObjStm")
	markerTypeObjStm    = []byte("/Type/ObjStm")
	markerFTBtn         = []byte("/FT /Btn")
	markerFTTx          = []byte("/FT /Tx")
	markerFTTxCompact   = []byte("/FT/Tx")
	markerFTBtnCompact  = []byte("/FT/Btn")
	markerParent        = []byte("/Parent")
	markerWidgetSubtype = []byte("/Subtype /Widget")
)

// Package-level compiled regexes (stable patterns used on hot paths / in loops).
var (
	valueTokenRe   = regexp.MustCompile(`/V\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	asTokenRe      = regexp.MustCompile(`/AS\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	arrayWRe       = regexp.MustCompile(`/W\s*\[(.*?)\]`)
	arrayIndexRe   = regexp.MustCompile(`/Index\s*\[(.*?)\]`)
	rootRefRe      = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	rootRef0Re     = regexp.MustCompile(`/Root\s+(\d+)\s+0\s+R`)
	acroFormRefRe  = regexp.MustCompile(`/AcroForm\s+(\d+)\s+(\d+)\s+R`)
	parenStrRe     = regexp.MustCompile(`\(([^)]{1,200})\)`)
	hexStrRe       = regexp.MustCompile(`<([0-9A-Fa-f\s]{2,400})>`)
	nameTokenRe    = regexp.MustCompile(`/([A-Za-z0-9_+-]{1,200})`)
	tFieldRe       = regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	tNameSimpleRe  = regexp.MustCompile(`/T\s*\((.*?)\)`)
	tNameAltRe     = regexp.MustCompile(`/T\s*(?:\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	kidsArrayRe    = regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	refRe          = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	singleKidsRe   = regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	vRefRe         = regexp.MustCompile(`/V\s*(\d+)\s+(\d+)\s+R`)
	streamNLRe     = regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	streamFlexRe   = regexp.MustCompile(`(?s)stream[\r\n]+(.*?)(?:[\r\n]+endstream|endstream)`)
	widgetASRe     = regexp.MustCompile(`/AS\s*(/(\w+)|\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
	apDictInnerRe  = regexp.MustCompile(`/AP\s*<<(.*?)>>`)
	nDictInnerRe   = regexp.MustCompile(`/N\s*<<(.*?)>>`)
	apNKeyRe       = regexp.MustCompile(`/([A-Za-z0-9_+-]+)\s*(?:/|stream|<<|\()`)
	nNameRe        = regexp.MustCompile(`/N\s*/([A-Za-z0-9_+-]+)`)
	trailerDictRe  = regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	objStreamRe    = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	objAtOffsetRe  = regexp.MustCompile(`(?s)^(\s*)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	firstNumRe     = regexp.MustCompile(`/First\s+(\d+)`)
	fieldsArrayRe  = regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
	singleFieldsRe = regexp.MustCompile(`/Fields\s+(\d+)\s+(\d+)\s+R`)
	widgetDictRe   = regexp.MustCompile(`(?s)<<.*?/Subtype\s*/Widget.*?>>`)
	apNExportRe    = regexp.MustCompile(`/AP\s*<<.*?/N\s*<<\s*/\s*([A-Za-z0-9_]+)\s*`)
	rectArrayRe    = regexp.MustCompile(`/Rect\s*\[\s*([^\]]+)\s*\]`)
	qValueRe       = regexp.MustCompile(`/Q\s*(\d)`)
	daValueRe      = regexp.MustCompile(`/DA\s*\((.*?)\)`)
	tfValueRe      = regexp.MustCompile(`/([\w.-]+)\s+([\d.]+)\s+Tf`)
	vParenMaybeRe  = regexp.MustCompile(`/V\s*\(?.*?\)?`)
	vParenRe       = regexp.MustCompile(`/V\s*\(.*?\)`)
	apStripRe      = regexp.MustCompile(`\s*/AP\s*<<.*?>>`)
	asNameRe       = regexp.MustCompile(`/AS\s*/\w+`)
	obj0Re         = regexp.MustCompile(`(\d+)\s+0\s+obj`)
	needAppRe      = regexp.MustCompile(`/NeedAppearances\s+(true|false)`)
	acroFormOpenRe = regexp.MustCompile(`(/AcroForm\s*<<)`)
	acroFormAnyRe  = regexp.MustCompile(`(?s)(/AcroForm\s*<<.*?)(>>)|(/AcroForm\s+\d+\s+\d+\s+R)`)
	apRefRe        = regexp.MustCompile(`/AP\s+\d+\s+\d+\s+R`)
	lengthNumRe    = regexp.MustCompile(`/Length\s+\d+`)

	// Precompiled patterns for replaceOrInsertPDFEntry
	vEntryTxRe  = regexp.MustCompile(`/V\s*\((?:\\.|[^\\)])*\)|/V\s*/[^\s/>]+|/V\s*<[0-9A-Fa-f\s]+>`)
	vEntryBtnRe = regexp.MustCompile(`/V\s*/[^\s/>]+|/V\s*\((?:\\.|[^\\)])*\)|/V\s*<[0-9A-Fa-f\s]+>`)
	asEntryRe   = regexp.MustCompile(`/AS\s*/[^\s/>]+|/AS\s*\((?:\\.|[^\\)])*\)|/AS\s*<[0-9A-Fa-f\s]+>`)
)

// fieldNameReCache caches dynamic field-name regexes (QuoteMeta patterns).
// Bounded to avoid unbounded growth (PERF-106/213).
const maxFieldNameReCache = 256

var (
	fieldNameReCache     sync.Map // string -> *regexp.Regexp
	fieldNameReCacheSize atomic.Int64
	fieldNameReCacheMu   sync.Mutex
)

func storeFieldNameRe(key string, re *regexp.Regexp) *regexp.Regexp {
	if v, loaded := fieldNameReCache.LoadOrStore(key, re); loaded {
		return v.(*regexp.Regexp)
	}
	n := fieldNameReCacheSize.Add(1)
	if n <= maxFieldNameReCache {
		return re
	}
	// Best-effort eviction of one other entry
	fieldNameReCacheMu.Lock()
	if fieldNameReCacheSize.Load() > maxFieldNameReCache {
		fieldNameReCache.Range(func(k, _ any) bool {
			if ks, ok := k.(string); ok && ks != key {
				fieldNameReCache.Delete(ks)
				fieldNameReCacheSize.Add(-1)
				return false
			}
			return true
		})
	}
	fieldNameReCacheMu.Unlock()
	return re
}

func loadFieldNameRe(key string) (*regexp.Regexp, bool) {
	v, ok := fieldNameReCache.Load(key)
	if !ok {
		return nil, false
	}
	re, ok := v.(*regexp.Regexp)
	return re, ok
}

func cachedFieldKidsDictRe(fieldName string) *regexp.Regexp {
	key := "kids:" + fieldName
	if re, ok := loadFieldNameRe(key); ok {
		return re
	}
	re := regexp.MustCompile(`(?s)(<<.*?/T\s*\(\s*` + regexp.QuoteMeta(fieldName) + `\s*\).*?/Kids.*?>>)`)
	return storeFieldNameRe(key, re)
}

func cachedFieldWidgetDictRe(fieldName string) *regexp.Regexp {
	key := "widget:" + fieldName
	if re, ok := loadFieldNameRe(key); ok {
		return re
	}
	re := regexp.MustCompile(`(?s)(<<.*?/Subtype\s*/Widget.*?/T\s*\(\s*` + regexp.QuoteMeta(fieldName) + `\s*\).*?>>)`)
	return storeFieldNameRe(key, re)
}

func cachedObjBodyRe(objNum, gen string) *regexp.Regexp {
	key := "obj:" + objNum + ":" + gen
	if re, ok := loadFieldNameRe(key); ok {
		return re
	}
	// objNum/gen come from digit captures; still QuoteMeta for safety
	re := regexp.MustCompile(`(?s)\b` + regexp.QuoteMeta(objNum) + `\s+` + regexp.QuoteMeta(gen) + `\s+obj(.*?)endobj`)
	return storeFieldNameRe(key, re)
}

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

	if m := valueTokenRe.FindSubmatch(window); m != nil {
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
	if m := asTokenRe.FindSubmatch(window); m != nil {
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
	var re *regexp.Regexp
	switch key {
	case `/W`:
		re = arrayWRe
	case `/Index`:
		re = arrayIndexRe
	default:
		re = regexp.MustCompile(regexp.QuoteMeta(key) + `\s*\[(.*?)\]`)
	}
	if m := re.FindSubmatch(dict); m != nil {
		inner := strings.TrimSpace(string(m[1]))
		if inner == "" {
			return nil
		}
		// PERF-186: single-pass whitespace scan instead of Fields
		return parseWhitespaceInts(inner)
	}
	return nil
}

// parseWhitespaceInts parses space-separated integers without strings.Fields.
func parseWhitespaceInts(s string) []int {
	res := make([]int, 0, 8)
	i := 0
	n := len(s)
	for i < n {
		for i < n && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
			i++
		}
		if i >= n {
			break
		}
		start := i
		for i < n && s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			i++
		}
		if v, err := strconv.Atoi(s[start:i]); err == nil {
			res = append(res, v)
		}
	}
	return res
}

// parseWhitespaceFloats parses up to max space-separated floats without Fields.
func parseWhitespaceFloats(s string, max int) []float64 {
	res := make([]float64, 0, max)
	i := 0
	n := len(s)
	for i < n && len(res) < max {
		for i < n && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
			i++
		}
		if i >= n {
			break
		}
		start := i
		for i < n && s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			i++
		}
		if v, err := strconv.ParseFloat(s[start:i], 64); err == nil {
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
	if m := rootRefRe.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// getAcroFormRef finds /AcroForm n m R reference
func getAcroFormRef(body []byte, data []byte) (string, bool) {
	if m := acroFormRefRe.FindSubmatch(body); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	if m := acroFormRefRe.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// extractStringFromBytes looks for PDF literal representations
func extractStringFromBytes(b []byte) string {
	if m := parenStrRe.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	if m := hexStrRe.FindSubmatch(b); m != nil {
		return decodeHexString(string(m[1]))
	}
	if m := nameTokenRe.FindSubmatch(b); m != nil {
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
	if m := tFieldRe.FindSubmatchIndex(body); m != nil {
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

	if m := kidsArrayRe.FindSubmatch(body); m != nil {
		inner := m[1]
		for _, r := range refRe.FindAllSubmatch(inner, -1) {
			kidRef := string(r[1]) + " " + string(r[2])
			traverseField(kidRef, objMap, fullName, out)
		}
	}
	if m := singleKidsRe.FindSubmatch(body); m != nil {
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
		if m := vRefRe.FindSubmatch(b); m != nil {
			ref := string(m[1]) + " " + string(m[2])
			if rb, ok := objMap[ref]; ok {
				if sm := streamNLRe.FindSubmatch(rb); sm != nil {
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
		if bytesIndex(body, markerSubtypeWidget) < 0 {
			continue
		}
		if bytesIndex(body, needle) < 0 && bytesIndex(body, markerT) < 0 {
			continue
		}
		if bytesIndex(body, needle) >= 0 {
			if m := widgetASRe.FindSubmatch(body); m != nil {
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
			if am := apDictInnerRe.FindSubmatch(body); am != nil {
				if nm := nDictInnerRe.FindSubmatch(am[1]); nm != nil {
					if kr := apNKeyRe.FindSubmatch(nm[1]); kr != nil {
						return string(kr[1]), true
					}
				}
				if nn := nNameRe.FindSubmatch(am[1]); nn != nil {
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
	for _, m := range trailerDictRe.FindAllSubmatch(data, -1) {
		if bytesIndex(m[1], markerEncrypt) >= 0 {
			return true
		}
	}
	// also check for /Encrypt elsewhere
	return bytesIndex(data, markerEncrypt) >= 0
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap
func parseXRefStreams(data []byte, objMap map[string][]byte) {
	// find objects with streams that contain /W and /Index
	for _, m := range objStreamRe.FindAllSubmatch(data, -1) {
		body := m[3]
		if bytesIndex(body, markerWBracket) < 0 || bytesIndex(body, markerIndex) < 0 {
			continue
		}
		// extract stream
		sm := streamFlexRe.FindSubmatch(body)
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
					if ro := objAtOffsetRe.FindSubmatch(tail); ro != nil {
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
				var tmp [20]byte
				key := string(strconv.AppendInt(tmp[:0], int64(objstm), 10)) + " 0"
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
	objMatches := objStreamRe.FindAllSubmatch(pdfBytes, -1)

	if len(objMatches) == 0 {
		// Fall back to naive scan
		return detectFormFieldsNaive(pdfBytes)
	}

	objMap := make(map[string][]byte, len(objMatches))
	for _, m := range objMatches {
		key := string(m[1]) + " " + string(m[2])
		body := m[3]

		// Handle ObjStm objects
		if bytesIndex(body, markerObjStm) >= 0 || bytesIndex(body, markerTypeObjStm) >= 0 {
			// find stream
			if sm := streamFlexRe.FindSubmatch(body); sm != nil {
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
					if fm := firstNumRe.FindSubmatch(body); fm != nil {
						if n, err := strconv.Atoi(string(fm[1])); err == nil {
							first = n
						}
					}
					if first > 0 && first < len(dec) {
						// parse header portion up to first (PERF-186: no Fields)
						header := strings.TrimSpace(string(dec[:first]))
						headerNums := parseWhitespaceInts(header)
						// header should be pairs of (objnum offset)
						pairs := make([][]int, 0, len(headerNums)/2)
						for i := 0; i+1 < len(headerNums); i += 2 {
							pairs = append(pairs, []int{headerNums[i], headerNums[i+1]})
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
							var keyTmp [20]byte
							objKey := string(strconv.AppendInt(keyTmp[:0], int64(objnum), 10)) + " 0"
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
	structured := make(map[string]string, 64) // PERF-192
	if rootRef, ok := findRootRef(pdfBytes); ok {
		if rootBody, ok2 := objMap[rootRef]; ok2 {
			if acroRef, ok3 := getAcroFormRef(rootBody, pdfBytes); ok3 {
				if afBody, ok4 := objMap[acroRef]; ok4 {
					if fm := fieldsArrayRe.FindSubmatch(afBody); fm != nil {
						inner := fm[1]
						for _, r := range refRe.FindAllSubmatch(inner, -1) {
							fref := string(r[1]) + " " + string(r[2])
							traverseField(fref, objMap, "", structured)
						}
					} else {
						if sm := singleFieldsRe.FindSubmatch(afBody); sm != nil {
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
	matches := tFieldRe.FindAllSubmatchIndex(pdfBytes, -1)

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
		for _, sm := range streamFlexRe.FindAllSubmatchIndex(newBody, -1) {
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

	mergedFields := make(map[string]string, len(detectedFields)+len(xfdfFields)) // PERF-192
	// maps.Copy avoids manual range loops flagged as PERF-114
	maps.Copy(mergedFields, detectedFields)
	maps.Copy(mergedFields, xfdfFields)

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

	out := bytes.Clone(pdfBytes)

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
	widgetMatches := widgetDictRe.FindAllIndex(out, -1)

	discoveredWidgets := make(map[string][]job, len(fields))

	for _, match := range widgetMatches {
		dictStart, dictEnd := match[0], match[1]
		dictBytes := out[dictStart:dictEnd]
		nameMatch := tNameSimpleRe.FindSubmatch(dictBytes)
		if nameMatch == nil {
			continue
		}
		fieldName := string(nameMatch[1])

		if _, ok := fields[fieldName]; !ok {
			continue // Skip if this field is not in our input data
		}

		currentVal := fields[fieldName]
		newJob := job{field: fieldName, val: currentVal, dictStart: dictStart, dictEnd: dictEnd}

		if bytes.Contains(dictBytes, markerFTBtn) {
			if bytes.Contains(dictBytes, markerParent) {
				newJob.fieldType = typeRadio
				if apMatch := apNExportRe.FindSubmatch(dictBytes); apMatch != nil {
					newJob.radioExportValue = string(apMatch[1])
				}
			} else {
				newJob.fieldType = typeButton
			}
		} else if bytes.Contains(dictBytes, markerFTTx) {
			newJob.fieldType = typeText
			rectMatch := rectArrayRe.FindSubmatch(dictBytes)
			if rectMatch == nil {
				continue
			}
			// PERF-186: parse rect coords without Fields allocation
			coords := parseWhitespaceFloats(string(rectMatch[1]), 4)
			if len(coords) < 4 {
				continue
			}
			llx, lly, urx, ury := coords[0], coords[1], coords[2], coords[3]
			newJob.width, newJob.height = urx-llx, ury-lly

			if qMatch := qValueRe.FindSubmatch(dictBytes); qMatch != nil {
				newJob.q, _ = strconv.Atoi(string(qMatch[1]))
			}
			newJob.fontSize, newJob.fontResourceName = 12.0, "Helv" // Default to Helv
			if daMatch := daValueRe.FindSubmatch(dictBytes); daMatch != nil {
				if tfMatch := tfValueRe.FindStringSubmatch(string(daMatch[1])); len(tfMatch) > 2 {
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

		radioGroups := make(map[string]string, 8) // PERF-192
		for _, job := range allJobs {
			if job.fieldType == typeRadio {
				radioGroups[job.field] = job.val
			}
		}
		maxVal := 0
		for _, value := range radioGroups {
			if len(value) > maxVal {
				maxVal = len(value)
			}
		}
		newV := make([]byte, 4+maxVal) // PERF-3: allocate once before loop
		for fieldName, value := range radioGroups {
			re := cachedFieldKidsDictRe(fieldName)
			match := re.FindIndex(out)
			if match == nil || bytes.Contains(out[match[0]:match[1]], markerWidgetSubtype) {
				continue
			}
			dictStart, dictEnd := match[0], match[1]
			dictBytes := out[dictStart:dictEnd]
			newV = newV[:4+len(value)]
			copy(newV, "/V /")
			copy(newV[4:], value)
			var newDictBytes []byte
			if vParenMaybeRe.Match(dictBytes) {
				newDictBytes = vParenMaybeRe.ReplaceAll(dictBytes, newV)
			} else {
				newDictBytes = bytes.Replace(dictBytes, []byte(">>"), append([]byte(" "), append(newV, []byte(">>")...)...), 1)
			}
			// PERF-119
			spliced := make([]byte, 0, len(out)-(dictEnd-dictStart)+len(newDictBytes))
			out = append(append(append(spliced, out[:dictStart]...), newDictBytes...), out[dictEnd:]...)
		}

		for _, job := range allJobs {
			dictBytes := out[job.dictStart:job.dictEnd]
			var newDictBytes []byte
			switch job.fieldType {
			case typeText:
				esc := escapePDFString(job.val)
				newV := []byte("/V (" + esc + ")")
				if vParenRe.Match(dictBytes) {
					newDictBytes = vParenRe.ReplaceAll(dictBytes, newV)
				} else {
					newDictBytes = bytes.Replace(dictBytes, []byte(">>"), append([]byte(" "), append(newV, []byte(">>")...)...), 1)
				}
				newDictBytes = apStripRe.ReplaceAll(newDictBytes, []byte(" "))
			case typeButton, typeRadio:
				newState := "/Off"
				v := job.val
				isButton := job.fieldType == typeButton
				isYesOn := len(v) == 2 && (v[0]|0x20) == 'o' && (v[1]|0x20) == 'n' ||
					len(v) == 3 && (v[0]|0x20) == 'y' && (v[1]|0x20) == 'e' && (v[2]|0x20) == 's'
				if isButton && isYesOn {
					if apMatch := apNExportRe.FindSubmatch(dictBytes); apMatch != nil {
						newState = "/" + string(apMatch[1])
					} else {
						newState = "/Yes"
					}
				} else if job.fieldType == typeRadio && job.radioExportValue == job.val {
					newState = "/" + job.radioExportValue
				}
				newAS := []byte("/AS " + newState)
				if asNameRe.Match(dictBytes) {
					newDictBytes = asNameRe.ReplaceAll(dictBytes, newAS)
				} else {
					newDictBytes = bytes.Replace(dictBytes, []byte(">>"), append([]byte(" "), append(newAS, []byte(">>")...)...), 1)
				}
			}
			if newDictBytes != nil {
				// PERF-119
				spliced := make([]byte, 0, len(out)-(job.dictEnd-job.dictStart)+len(newDictBytes))
				out = append(append(append(spliced, out[:job.dictStart]...), newDictBytes...), out[job.dictEnd:]...)
			}
		}
	}

	// --- PASS 3: NEW OBJECT GENERATION ---
	allObjMatches := obj0Re.FindAllSubmatchIndex(out, -1)
	highest := 0
	for _, m := range allObjMatches {
		if n, err := strconv.Atoi(string(out[m[2]:m[3]])); err == nil && n > highest {
			highest = n
		}
	}
	nextObj := highest + 1
	if sx := bytes.LastIndex(out, []byte("startxref")); sx >= 0 {
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
		re := cachedFieldWidgetDictRe(job.field)
		match := re.FindIndex(out)
		if match == nil {
			continue
		}
		dictEnd := match[1]
		var apTmp [20]byte
		apRef := append(append([]byte(" /AP<</N "), strconv.AppendInt(apTmp[:0], int64(job.apObjNum), 10)...), " 0 R>>"...)
		// PERF-119: avoid nested append(a, append(b, c...)...)
		newOut := make([]byte, 0, len(out)+len(apRef))
		newOut = append(append(append(newOut, out[:dictEnd-2]...), apRef...), out[dictEnd-2:]...)
		out = newOut
	}

	if needAppRe.Match(out) {
		out = needAppRe.ReplaceAll(out, []byte("/NeedAppearances false"))
	} else {
		if loc := acroFormOpenRe.FindIndex(out); loc != nil {
			insertPos := loc[1]
			insertContent := []byte(" /NeedAppearances false ")
			// PERF-119: pre-size once for three-segment splice
			newOut := make([]byte, 0, len(out)+len(insertContent))
			newOut = append(append(append(newOut, out[:insertPos]...), insertContent...), out[insertPos:]...)
			out = newOut
		}
	}

	// Prebuild Helvetica widths array once (PERF-6: outside textJobs loop)
	var widthsBuf strings.Builder
	var widthTmp [20]byte
	widthsBuf.Grow(len(helveticaWidths)*4 + 2)
	widthsBuf.WriteByte('[')
	for i, w := range helveticaWidths {
		if i > 0 {
			widthsBuf.WriteByte(' ')
		}
		widthsBuf.Write(strconv.AppendInt(widthTmp[:0], int64(w), 10))
	}
	widthsBuf.WriteByte(']')
	widthsStr := widthsBuf.String()

	var intTmp [20]byte
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
		// PERF-6/15: build appearance stream with AppendFloat (no FormatFloat allocs)
		var floatTmp [32]byte
		var streamSB strings.Builder
		streamSB.Grow(64 + len(streamText))
		streamSB.WriteString("q\nBT\n/F1 ")
		streamSB.Write(strconv.AppendFloat(floatTmp[:0], job.fontSize, 'f', 2, 64))
		streamSB.WriteString(" Tf\n0 g\n")
		streamSB.Write(strconv.AppendFloat(floatTmp[:0], tx, 'f', 2, 64))
		streamSB.WriteByte(' ')
		streamSB.Write(strconv.AppendFloat(floatTmp[:0], y, 'f', 2, 64))
		streamSB.WriteString(" Td\n(")
		streamSB.WriteString(streamText)
		streamSB.WriteString(") Tj\nET\nQ")
		streamBody := streamSB.String()

		var fontDescSB strings.Builder
		fontDescSB.Grow(200 + len(job.fontResourceName))
		fontDescSB.WriteByte('\n')
		fontDescSB.Write(strconv.AppendInt(intTmp[:0], int64(job.fontDescObjNum), 10))
		fontDescSB.WriteString(" 0 obj\n<</Type/FontDescriptor/FontName/")
		fontDescSB.WriteString(job.fontResourceName)
		fontDescSB.WriteString("/Flags 32/FontBBox[-558 -225 1000 931]/ItalicAngle 0/Ascent 905/Descent -212/CapHeight 905/StemV 88>>\nendobj\n")
		out = append(out, fontDescSB.String()...)

		// Update the Font object to include FirstChar, LastChar, and the Widths array.
		// Using full WinAnsiEncoding range (32-255) for PDF 2.0 compliance
		var fontSB strings.Builder
		fontSB.Grow(200 + len(job.fontResourceName) + len(widthsStr))
		fontSB.WriteByte('\n')
		fontSB.Write(strconv.AppendInt(intTmp[:0], int64(job.fontObjNum), 10))
		fontSB.WriteString(" 0 obj\n<</Type/Font/Subtype/Type1/BaseFont/")
		fontSB.WriteString(job.fontResourceName)
		fontSB.WriteString("/Encoding/WinAnsiEncoding/FirstChar 32/LastChar 255/Widths ")
		fontSB.WriteString(widthsStr)
		fontSB.WriteString("/FontDescriptor ")
		fontSB.Write(strconv.AppendInt(intTmp[:0], int64(job.fontDescObjNum), 10))
		fontSB.WriteString(" 0 R>>\nendobj\n")
		out = append(out, fontSB.String()...)

		var compBuf bytes.Buffer
		zw, _ := zlib.NewWriterLevel(&compBuf, zlib.BestSpeed)
		if _, err := zw.Write(byteconv.StringToBytes(streamBody)); err != nil {
			return nil, errors.Join(errors.New("compression write failed"), err)
		}
		if err := zw.Close(); err != nil {
			return nil, errors.Join(errors.New("compression close failed"), err)
		}
		comp := compBuf.Bytes()
		var apSB strings.Builder
		apSB.Grow(160 + len(comp))
		apSB.WriteByte('\n')
		apSB.Write(strconv.AppendInt(intTmp[:0], int64(job.apObjNum), 10))
		apSB.WriteString(" 0 obj\n<</Type/XObject/Subtype/Form/FormType 1/BBox[0 0 ")
		apSB.Write(strconv.AppendFloat(floatTmp[:0], job.width, 'f', 2, 64))
		apSB.WriteByte(' ')
		apSB.Write(strconv.AppendFloat(floatTmp[:0], job.height, 'f', 2, 64))
		apSB.WriteString("]/Resources<</Font<</F1 ")
		apSB.Write(strconv.AppendInt(intTmp[:0], int64(job.fontObjNum), 10))
		apSB.WriteString(" 0 R>>/ProcSet[/PDF/Text]>>/Filter/FlateDecode/Length ")
		apSB.Write(strconv.AppendInt(intTmp[:0], int64(len(comp)), 10))
		apSB.WriteString(">>\nstream\n")
		apSB.Write(comp)
		apSB.WriteString("\nendstream\nendobj\n")
		out = append(out, apSB.String()...)
	}

	objMatches := obj0Re.FindAllSubmatchIndex(out, -1)
	offsets := make(map[int]int, len(objMatches))
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
	var xrefTmp [20]byte
	xrefBuf.WriteString("xref\n0 ")
	xrefBuf.Write(strconv.AppendInt(xrefTmp[:0], int64(maxObj+1), 10))
	xrefBuf.WriteByte('\n')
	xrefBuf.WriteString("0000000000 65535 f \r\n")
	// Fixed 20-byte xref entry: 10 digit offset + " 00000 n \r\n" (PERF-119/128)
	var xrefEntry [20]byte
	copy(xrefEntry[10:], " 00000 n \r\n")
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			offStr := strconv.AppendInt(xrefTmp[:0], int64(off), 10)
			pad := 10 - len(offStr)
			for j := 0; j < pad; j++ {
				xrefEntry[j] = '0'
			}
			copy(xrefEntry[pad:10], offStr)
			xrefBuf.Write(xrefEntry[:])
		} else {
			xrefBuf.WriteString("0000000000 65535 f \r\n")
		}
	}
	root := 1
	if rm := rootRef0Re.FindSubmatch(pdfBytes); len(rm) > 1 {
		if r, err := strconv.Atoi(string(rm[1])); err == nil {
			root = r
		}
	}
	var trailerBuf strings.Builder
	trailerBuf.WriteString("trailer\n<</Size ")
	trailerBuf.Write(strconv.AppendInt(xrefTmp[:0], int64(maxObj+1), 10))
	trailerBuf.WriteString("/Root ")
	trailerBuf.Write(strconv.AppendInt(xrefTmp[:0], int64(root), 10))
	trailerBuf.WriteString(" 0 R>>\nstartxref\n")
	trailerBuf.Write(strconv.AppendInt(xrefTmp[:0], int64(xrefStart), 10))
	trailerBuf.WriteString("\n%%%%EOF\n")
	// PERF-119: one growth for xref + trailer
	xrefBytes := xrefBuf.Bytes()
	trailerBytes := []byte(trailerBuf.String())
	tail := make([]byte, 0, len(xrefBytes)+len(trailerBytes))
	tail = append(tail, xrefBytes...)
	tail = append(tail, trailerBytes...)
	out = append(out, tail...)
	// --- PASS 3: GLOBAL NEED APPEARANCES ---
	// If fields were modified or APs stripped, force the PDF viewer to recreate appearances on open.
	acroMatch := acroFormAnyRe.FindSubmatch(out)
	if acroMatch != nil {
		if acroMatch[1] != nil {
			// Inline dictionary case
			dictPart := acroMatch[1]
			if needAppRe.Match(dictPart) {
				newDict := needAppRe.ReplaceAll(dictPart, []byte("/NeedAppearances true"))
				out = bytes.Replace(out, dictPart, newDict, 1)
			} else {
				// Inject it
				newDict := make([]byte, 0, len(dictPart)+len(" /NeedAppearances true "))
				newDict = append(newDict, dictPart...)
				newDict = append(newDict, " /NeedAppearances true "...)
				out = bytes.Replace(out, dictPart, newDict, 1)
			}
		} else if acroMatch[3] != nil {
			// Indirect reference case
			refFull := string(acroMatch[3])
			if rm := refRe.FindStringSubmatch(refFull); rm != nil {
				objBodyRe := cachedObjBodyRe(rm[1], rm[2])
				if objM := objBodyRe.FindSubmatch(out); objM != nil {
					objBody := objM[1]
					if needAppRe.Match(objBody) {
						newBody := needAppRe.ReplaceAll(objBody, []byte("/NeedAppearances true"))
						out = bytes.Replace(out, objBody, newBody, 1)
					} else {
						// Inject before the ending >>
						insertPos := bytes.LastIndex(objBody, []byte(">>"))
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
	out := bytes.Clone(pdfBytes)

	matches := objStreamRe.FindAllSubmatchIndex(out, -1)
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
		// PERF-119
		spliced := make([]byte, 0, len(out)-(bodyEnd-bodyStart)+len(newBody))
		out = append(append(append(spliced, out[:bodyStart]...), newBody...), out[bodyEnd:]...)
	}

	return out, changedAny, nil
}

//nolint:gocyclo
func fillXFDFInObjStmBody(body []byte, fields map[string]string) ([]byte, bool, error) {
	sm := streamFlexRe.FindSubmatchIndex(body)
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

	fm := firstNumRe.FindSubmatch(body)
	if fm == nil {
		return body, false, nil
	}
	first, err := strconv.Atoi(string(fm[1]))
	if err != nil || first <= 0 || first > len(decoded) {
		return body, false, nil
	}

	header := strings.TrimSpace(string(decoded[:first]))
	// PERF-186: parse ObjStm header number pairs without Fields
	headerNums := parseWhitespaceInts(header)
	if len(headerNums) < 2 || len(headerNums)%2 != 0 {
		return body, false, nil
	}

	type objMember struct {
		objNum  int
		offset  int
		content []byte
	}

	content := decoded[first:]
	members := make([]objMember, 0, len(headerNums)/2)
	for i := 0; i < len(headerNums); i += 2 {
		members = append(members, objMember{objNum: headerNums[i], offset: headerNums[i+1]})
	}

	kidsToRemoveAP := make(map[int]bool, len(members)) // PERF-192
	for i := range members {
		start := members[i].offset
		end := len(content)
		if i+1 < len(members) {
			end = members[i+1].offset
		}
		objContent := bytes.TrimSpace(content[start:end])

		if nameMatch := tNameAltRe.FindSubmatch(objContent); nameMatch != nil {
			var fieldName string
			if len(nameMatch[1]) > 0 {
				fieldName = string(nameMatch[1])
			} else if len(nameMatch[2]) > 0 {
				fieldName = decodeHexString(string(nameMatch[2]))
			}
			fieldName = strings.TrimSpace(fieldName)

			if _, ok := fields[fieldName]; ok {
				if bytes.Contains(objContent, markerFTTx) || bytes.Contains(objContent, markerFTTxCompact) {
					kidsToRemoveAP[members[i].objNum] = true
					if m := kidsArrayRe.FindSubmatch(objContent); m != nil {
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
			if apRefRe.Match(updated) {
				updated = apRefRe.ReplaceAll(updated, []byte(" "))
				changed = true
			}
			// Manual removal of /AP dictionary to handle nested <<>>
			apIdx := bytes.Index(updated, []byte("/AP"))
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
		if bytes.Contains(updated, []byte("/Fields[")) || bytes.Contains(updated, []byte("/Fields [")) {
			// Basic AcroForm object identification
			if !bytes.Contains(updated, []byte("/NeedAppearances")) {
				insertPos := bytes.LastIndex(updated, []byte(">>"))
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
	var hdrTmp [20]byte
	currentOffset := 0
	for i, member := range members {
		headerBuilder.Write(strconv.AppendInt(hdrTmp[:0], int64(member.objNum), 10))
		headerBuilder.WriteByte(' ')
		headerBuilder.Write(strconv.AppendInt(hdrTmp[:0], int64(currentOffset), 10))
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
	zw, err := zlib.NewWriterLevel(&compressedBuf, zlib.BestSpeed)
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
	var dictTmp [20]byte
	firstRepl := append([]byte("/First "), strconv.AppendInt(dictTmp[:0], int64(newFirst), 10)...)
	newDict := firstNumRe.ReplaceAll(dictPart, firstRepl)
	if lengthNumRe.Match(newDict) {
		lenRepl := append([]byte("/Length "), strconv.AppendInt(dictTmp[:0], int64(len(compressed)), 10)...)
		newDict = lengthNumRe.ReplaceAll(newDict, lenRepl)
	}

	var rebuilt bytes.Buffer
	rebuilt.Grow(len(newDict) + len("stream\n") + len(compressed) + len("\nendstream") + len(suffix))
	rebuilt.Write(newDict)
	rebuilt.WriteString("stream\n")
	rebuilt.Write(compressed)
	rebuilt.WriteString("\nendstream")
	rebuilt.Write(suffix)

	return rebuilt.Bytes(), true, nil
}

func updateObjStmFieldValue(objContent []byte, fields map[string]string) ([]byte, bool) {
	nameMatch := tNameAltRe.FindSubmatch(objContent)
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

	updated := bytes.Clone(objContent)

	if bytes.Contains(updated, markerFTTx) || bytes.Contains(updated, markerFTTxCompact) {
		replacement := []byte("/V (" + escapePDFString(value) + ")")
		newUpdated, changed := replaceOrInsertPDFEntry(updated, vEntryTxRe, replacement)
		return newUpdated, changed
	}

	if bytes.Contains(updated, markerFTBtn) || bytes.Contains(updated, markerFTBtnCompact) {
		state := xfdfValueToPDFName(value)
		newUpdated, changedV := replaceOrInsertPDFEntry(updated, vEntryBtnRe, []byte("/V /"+state))
		newUpdated, changedAS := replaceOrInsertPDFEntry(newUpdated, asEntryRe, []byte("/AS /"+state))
		return newUpdated, changedV || changedAS
	}

	return objContent, false
}

func replaceOrInsertPDFEntry(dict []byte, re *regexp.Regexp, replacement []byte) ([]byte, bool) {
	if re.Match(dict) {
		newDict := re.ReplaceAll(dict, replacement)
		return newDict, !bytes.Equal(newDict, dict)
	}

	insertPos := bytes.LastIndex(dict, []byte(">>"))
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
