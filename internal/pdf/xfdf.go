package pdf

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

	"github.com/chinmay-sawant/gopdfsuit/scripts"
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
	m := make(map[string]string)
	for _, f := range root.Fields {
		name := strings.TrimSpace(f.Name)
		val := strings.TrimSpace(f.Value)
		// Don't escape the XFDF value here - keep it as original
		m[name] = val
	}
	return m, nil
}

// Enhanced field detection structures
type FormField struct {
	Name  string
	Value string
	Type  string // V, AS, or detected type
}

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return strings.Index(string(b), string(sub))
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
	defer r.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// tryFlateDecompress attempts to decompress raw flate data
func tryFlateDecompress(b []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(b))
	defer r.Close()
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

	valueRe := regexp.MustCompile(`/V\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	asRe := regexp.MustCompile(`/AS\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)

	if m := valueRe.FindSubmatch(window); m != nil {
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
	if m := asRe.FindSubmatch(window); m != nil {
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
	parenRe := regexp.MustCompile(`\(([^)]{1,200})\)`)
	if m := parenRe.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	hexRe := regexp.MustCompile(`<([0-9A-Fa-f\s]{2,400})>`)
	if m := hexRe.FindSubmatch(b); m != nil {
		return decodeHexString(string(m[1]))
	}
	nameRe := regexp.MustCompile(`/([A-Za-z0-9_+-]{1,200})`)
	if m := nameRe.FindSubmatch(b); m != nil {
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

	tReLocal := regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	tv := ""
	name := ""
	if m := tReLocal.FindSubmatchIndex(body); m != nil {
		if m[4] != -1 && m[5] != -1 {
			name = string(body[m[4]:m[5]])
		} else if m[6] != -1 && m[7] != -1 {
			name = decodeHexString(string(body[m[6]:m[7]]))
		} else if m[8] != -1 && m[9] != -1 {
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
		out[fullName] = tv
	}

	kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	if m := kidsRe.FindSubmatch(body); m != nil {
		inner := m[1]
		for _, r := range refRe.FindAllSubmatch(inner, -1) {
			kidRef := string(r[1]) + " " + string(r[2])
			traverseField(kidRef, objMap, fullName, out)
		}
	}
	singleKidsRe := regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
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
		refRe := regexp.MustCompile(`/V\s*(\d+)\s+(\d+)\s+R`)
		if m := refRe.FindSubmatch(b); m != nil {
			ref := string(m[1]) + " " + string(m[2])
			if rb, ok := objMap[ref]; ok {
				streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
				if sm := streamRe.FindSubmatch(rb); sm != nil {
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
		if bytesIndex(body, []byte(`/Subtype/Widget`)) < 0 {
			continue
		}
		if bytesIndex(body, needle) < 0 && bytesIndex(body, []byte(`/T`)) < 0 {
			continue
		}
		if bytesIndex(body, needle) >= 0 {
			asRe := regexp.MustCompile(`/AS\s*(/(\w+)|\(([^)]*)\)|<([0-9A-Fa-f\s]+)>)`)
			if m := asRe.FindSubmatch(body); m != nil {
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
			apRe := regexp.MustCompile(`/AP\s*<<(.*?)>>`)
			if am := apRe.FindSubmatch(body); am != nil {
				nRe := regexp.MustCompile(`/N\s*<<(.*?)>>`)
				if nm := nRe.FindSubmatch(am[1]); nm != nil {
					keyRe := regexp.MustCompile(`/([A-Za-z0-9_+-]+)\s*(?:/|stream|<<|\()`)
					if kr := keyRe.FindSubmatch(nm[1]); kr != nil {
						return string(kr[1]), true
					}
				}
				nNameRe := regexp.MustCompile(`/N\s*/([A-Za-z0-9_+-]+)`)
				if nn := nNameRe.FindSubmatch(am[1]); nn != nil {
					return string(nn[1]), true
				}
			}
			return k, true
		}
	}
	return "", false
}

// findStartXref finds the last startxref offset in the file
func findStartXref(data []byte) (int64, bool) {
	re := regexp.MustCompile(`startxref\s*(\d+)`)
	ms := re.FindAllSubmatch(data, -1)
	if len(ms) == 0 {
		return 0, false
	}
	last := ms[len(ms)-1]
	var off int64
	fmt.Sscanf(string(last[1]), "%d", &off)
	return off, true
}

// trailerHasEncrypt checks if trailer or any trailer 'Encrypt' appears
func trailerHasEncrypt(data []byte) bool {
	trRe := regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	for _, m := range trRe.FindAllSubmatch(data, -1) {
		if bytesIndex(m[1], []byte(`/Encrypt`)) >= 0 {
			return true
		}
	}
	// also check for /Encrypt elsewhere
	return bytesIndex(data, []byte(`/Encrypt`)) >= 0
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap
func parseXRefStreams(data []byte, objMap map[string][]byte) {
	// find objects with streams that contain /W and /Index
	objStreamRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	for _, m := range objStreamRe.FindAllSubmatch(data, -1) {
		body := m[3]
		if bytesIndex(body, []byte(`/W[`)) < 0 || bytesIndex(body, []byte(`/Index`)) < 0 {
			continue
		}
		// extract stream
		streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
		sm := streamRe.FindSubmatch(body)
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
					reObj := regexp.MustCompile(`(?s)^(\s*)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
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
				key := fmt.Sprintf("%d 0", objstm)
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
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	objMatches := objRe.FindAllSubmatch(pdfBytes, -1)

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
			streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
			if sm := streamRe.FindSubmatch(body); sm != nil {
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
					firstRe := regexp.MustCompile(`/First\s+(\d+)`)
					first := 0
					if fm := firstRe.FindSubmatch(body); fm != nil {
						fmt.Sscanf(string(fm[1]), "%d", &first)
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
							objKey := fmt.Sprintf("%d 0", objnum)
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
	tRe := regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
	matches := tRe.FindAllSubmatchIndex(pdfBytes, -1)

	result := make(map[string]string)
	seen := make(map[string]bool)

	for _, mi := range matches {
		var name string
		if mi[4] != -1 && mi[5] != -1 {
			name = string(pdfBytes[mi[4]:mi[5]])
		} else if mi[6] != -1 && mi[7] != -1 {
			name = decodeHexString(string(pdfBytes[mi[6]:mi[7]]))
		} else if mi[8] != -1 && mi[9] != -1 {
			name = string(pdfBytes[mi[8]:mi[9]])
		} else {
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
	streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
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

	// Parse XFDF data
	xfdfFields, err := ParseXFDF(xfdfBytes)
	if err != nil {
		return nil, err
	}

	// Detect existing fields in PDF
	detectedFields, err := DetectFormFieldsAdvanced(pdfBytes)
	if err != nil {
		return nil, err
	}

	// Merge XFDF data with detected fields
	mergedFields := make(map[string]string)
	for name, value := range detectedFields {
		mergedFields[name] = value
	}
	for name, value := range xfdfFields {
		mergedFields[name] = value // XFDF values override detected values
	}

	// Use existing value setting logic
	return FillPDFWithXFDF(pdfBytes, xfdfBytes)
}

// FillPDFWithXFDF attempts a best-effort in-place fill of PDF form fields using XFDF data.
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

	type job struct {
		field              string
		text               string
		dictStart          int
		dictEnd            int
		llx, lly, urx, ury float64
		width, height      float64
		q                  int
		fontSize           float64
		apObjNum           int
		fontName           string
		isButton           bool
		fontResourceName   string
	}
	jobs := make([]job, 0)

	for name, val := range fields {
		re := regexp.MustCompile(fmt.Sprintf(`/T\s*\(%s\)`, regexp.QuoteMeta(name)))
		matches := re.FindAllIndex(out, -1)
		if len(matches) == 0 {
			re2 := regexp.MustCompile(`(?i)/T\s*\(` + regexp.QuoteMeta(name) + `\)`)
			matches = re2.FindAllIndex(out, -1)
		}
		if len(matches) == 0 {
			continue
		}

		// Only escape for PDF syntax when setting the /V value
		esc := escapePDFString(val)
		newV := []byte(fmt.Sprintf("/V (%s)", esc))

		for i := len(matches) - 1; i >= 0; i-- {
			idx := matches[i][0]
			searchStart := idx
			searchEnd := idx + 2048
			if searchEnd > len(out) {
				searchEnd = len(out)
			}
			segment := out[searchStart:searchEnd]

			// Check if this is a button field
			isButton := bytes.Contains(segment, []byte("/FT /Btn"))

			// For button fields, handle both /V and /AS
			if isButton {
				// Handle /V field
				vRel := bytes.Index(segment, []byte("/V "))
				if vRel >= 0 {
					// Find the value - could be /V /Off, /V /Yes, or /V (value)
					vStart := searchStart + vRel
					vEnd := vStart + 3 // Start after "/V "

					// Skip whitespace
					for vEnd < len(out) && (out[vEnd] == ' ' || out[vEnd] == '\t' || out[vEnd] == '\n' || out[vEnd] == '\r') {
						vEnd++
					}

					if vEnd < len(out) {
						if out[vEnd] == '/' {
							// Name object like /Off or /Yes
							nameEnd := vEnd + 1
							for nameEnd < len(out) && out[nameEnd] != ' ' && out[nameEnd] != '\t' && out[nameEnd] != '\n' && out[nameEnd] != '\r' && out[nameEnd] != '>' {
								nameEnd++
							}
							// Replace with new value
							if strings.ToLower(val) == "yes" || strings.ToLower(val) == "true" || val == "1" {
								newButtonV := []byte("/V /Yes")
								out = append(out[:vStart], append(newButtonV, out[nameEnd:]...)...)
							} else {
								newButtonV := []byte("/V /Off")
								out = append(out[:vStart], append(newButtonV, out[nameEnd:]...)...)
							}
						} else if out[vEnd] == '(' {
							// String value
							valStart := vEnd + 1
							valEnd := bytes.IndexByte(out[valStart:], ')')
							if valEnd >= 0 {
								valEnd = valStart + valEnd
								out = append(out[:vStart], append(newV, out[valEnd+1:]...)...)
							}
						}
					}
				} else {
					// Insert /V if not found
					closerRel := bytes.Index(segment, []byte(">>"))
					if closerRel >= 0 {
						insertPos := searchStart + closerRel
						if strings.ToLower(val) == "yes" || strings.ToLower(val) == "true" || val == "1" {
							insertion := []byte(" /V /Yes")
							out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
						} else {
							insertion := []byte(" /V /Off")
							out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
						}
					}
				}

				// Handle /AS (Appearance State) field - this is crucial for Adobe Acrobat
				asRe := regexp.MustCompile(`/AS\s*/(\w+)`)
				if asMatch := asRe.FindIndex(segment); asMatch != nil {
					asStart := searchStart + asMatch[0]
					asEnd := searchStart + asMatch[1]

					if strings.ToLower(val) == "yes" || strings.ToLower(val) == "true" || val == "1" {
						newAS := []byte("/AS /Yes")
						out = append(out[:asStart], append(newAS, out[asEnd:]...)...)
					} else {
						newAS := []byte("/AS /Off")
						out = append(out[:asStart], append(newAS, out[asEnd:]...)...)
					}
				} else {
					// Insert /AS if not found
					closerRel := bytes.Index(segment, []byte(">>"))
					if closerRel >= 0 {
						insertPos := searchStart + closerRel
						if strings.ToLower(val) == "yes" || strings.ToLower(val) == "true" || val == "1" {
							insertion := []byte(" /AS /Yes")
							out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
						} else {
							insertion := []byte(" /AS /Off")
							out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
						}
					}
				}
			} else {
				// Handle text fields - set both /V value and create appearance
				vRel := bytes.Index(segment, []byte("/V ("))
				if vRel >= 0 {
					vStart := searchStart + vRel
					valStart := vStart + len([]byte("/V ("))
					valEnd := bytes.IndexByte(out[valStart:], ')')
					if valEnd < 0 {
						continue
					}
					valEnd = valStart + valEnd
					out = append(out[:vStart], append(newV, out[valEnd+1:]...)...)
				} else {
					closerRel := bytes.Index(segment, []byte(">>"))
					insertPos := -1
					if closerRel >= 0 {
						insertPos = searchStart + closerRel
					} else {
						endobjRel := bytes.Index(segment, []byte("endobj"))
						if endobjRel >= 0 {
							insertPos = searchStart + endobjRel
						}
					}
					if insertPos > 0 {
						insertion := append([]byte(" "), newV...)
						out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
					}
				}
			}

			// find surrounding dict for appearance streams (only for text fields)
			if !isButton {
				ctxStart := idx - 2048
				if ctxStart < 0 {
					ctxStart = 0
				}
				ctxEnd := idx + 2048
				if ctxEnd > len(out) {
					ctxEnd = len(out)
				}
				ctx := out[ctxStart:ctxEnd]
				relDictStart := bytes.LastIndex(ctx[:idx-ctxStart+1], []byte("<<"))
				if relDictStart < 0 {
					continue
				}
				dictStart := ctxStart + relDictStart

				depth := 0
				j := dictStart
				dictEnd := -1
				for j < ctxEnd-1 {
					if j+1 < len(out) && out[j] == '<' && out[j+1] == '<' {
						depth++
						j += 2
						continue
					}
					if j+1 < len(out) && out[j] == '>' && out[j+1] == '>' {
						depth--
						j += 2
						if depth == 0 {
							dictEnd = j - 2
							break
						}
						continue
					}
					j++
				}
				if dictEnd < 0 {
					continue
				}

				dictBytes := out[dictStart : dictEnd+2]

				// Remove existing /AP if present to replace it with our own
				apRe := regexp.MustCompile(`\s*/AP\s*<<[^>]*>>\s*`)
				if apRe.Match(dictBytes) {
					out = apRe.ReplaceAll(out, []byte(" "))
				}

				rectRe := regexp.MustCompile(`/Rect\s*\[\s*([^\]]+)\s*\]`)
				rectMatch := rectRe.FindSubmatch(dictBytes)
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
				width := urx - llx
				height := ury - lly

				q := 0
				qRe := regexp.MustCompile(`/Q\s*(\d)`)
				qMatch := qRe.FindSubmatch(dictBytes)
				if qMatch != nil {
					q, _ = strconv.Atoi(string(qMatch[1]))
				}

				fontSize := 12.0
				fontName := "Helv"
				fontResourceName := "Helv"
				daRe := regexp.MustCompile(`/DA\s*\(([^"]*)\)`)
				daMatch := daRe.FindSubmatch(dictBytes)
				if daMatch != nil {
					daStr := string(daMatch[1])
					tfRe := regexp.MustCompile(`/([\w]+)\s+(\d+\.?\d*)\s*Tf`)
					tfMatch := tfRe.FindStringSubmatch(daStr)
					if len(tfMatch) > 2 {
						fontName = tfMatch[1]
						fontResourceName = tfMatch[1]
						if fs, err := strconv.ParseFloat(tfMatch[2], 64); err == nil {
							fontSize = fs
						}
					}
				}

				// Auto-adjust font size if text is too large for field
				if len(val) > 0 {
					maxTextWidth := width - 6 // 3pt padding on each side
					estimatedTextWidth := float64(len(val)) * fontSize * 0.6
					if estimatedTextWidth > maxTextWidth && maxTextWidth > 0 {
						fontSize = maxTextWidth / (float64(len(val)) * 0.6)
						if fontSize < 8 {
							fontSize = 8 // Minimum readable font size
						}
					}
				}

				jobs = append(jobs, job{
					field:     name,
					text:      val, // Use original value without escaping for display
					dictStart: dictStart,
					dictEnd:   dictEnd,
					llx:       llx, lly: lly, urx: urx, ury: ury,
					width: width, height: height,
					q:                q,
					fontSize:         fontSize,
					fontName:         fontName,
					fontResourceName: fontResourceName,
					isButton:         false,
				})
			}
		}
	}

	if len(jobs) == 0 {
		// fallback: set NeedAppearances true
		acRe := regexp.MustCompile(`/AcroForm\s+(\d+)\s0\sR`)
		if am := acRe.FindSubmatch(pdfBytes); len(am) > 1 {
			objNum := string(am[1])
			objHeader := []byte(fmt.Sprintf("%s 0 obj", objNum))
			if objPos := bytes.Index(out, objHeader); objPos >= 0 {
				dictStartRel := bytes.Index(out[objPos:], []byte("<<"))
				if dictStartRel >= 0 {
					dictStart := objPos + dictStartRel
					depth := 0
					i := dictStart
					dictEnd := -1
					for i < len(out)-1 {
						if i+1 < len(out) && out[i] == '<' && out[i+1] == '<' {
							depth++
							i += 2
							continue
						}
						if i+1 < len(out) && out[i] == '>' && out[i+1] == '>' {
							depth--
							i += 2
							if depth == 0 {
								dictEnd = i - 2
								break
							}
							continue
						}
						i++
					}
					if dictEnd >= 0 {
						if !bytes.Contains(out[dictStart:dictEnd+2], []byte("/NeedAppearances")) {
							insertion := []byte(" /NeedAppearances true")
							out = append(out[:dictEnd], append(insertion, out[dictEnd:]...)...)
						}
					}
				}
			}
		}
		// After in-place edits, call FlattenPDFBytes to produce a flattened PDF bytes stream.
		// Note: FlattenPDFBytes is implemented in this package as a no-op stub currently.
		flat, err := FlattenPDFBytes(out)
		if err != nil {
			return nil, err
		}
		return flat, nil
	}

	// have jobs -> create AP objects
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].dictStart > jobs[j].dictStart })

	objRe := regexp.MustCompile(`(\d+)\s+0\s+obj`)
	allObjMatches := objRe.FindAllSubmatchIndex(out, -1)
	highest := 0
	for _, m := range allObjMatches {
		numBytes := out[m[2]:m[3]]
		if n, err := strconv.Atoi(string(numBytes)); err == nil && n > highest {
			highest = n
		}
	}
	nextObj := highest + 1

	if sx := bytes.LastIndex(out, []byte("startxref")); sx >= 0 {
		out = out[:sx]
	}

	delta := 0
	for i := range jobs {
		apNum := nextObj
		jobs[i].apObjNum = apNum
		insertPos := jobs[i].dictEnd + delta
		apRef := []byte(fmt.Sprintf(" /AP << /N %d 0 R >>", apNum))
		out = append(out[:insertPos], append(apRef, out[insertPos:]...)...)
		delta += len(apRef)
		nextObj++
	}

	// Always set NeedAppearances to false since we're providing proper appearances
	acRe := regexp.MustCompile(`/AcroForm\s+(\d+)\s0\sR`)
	if am := acRe.FindSubmatch(out); len(am) > 1 {
		objNum := string(am[1])
		objHeader := []byte(fmt.Sprintf("%s 0 obj", objNum))
		if objPos := bytes.Index(out, objHeader); objPos >= 0 {
			dictStartRel := bytes.Index(out[objPos:], []byte("<<"))
			if dictStartRel >= 0 {
				dictStart := objPos + dictStartRel
				depth := 0
				i := dictStart
				dictEnd := -1
				for i < len(out)-1 {
					if i+1 < len(out) && out[i] == '<' && out[i+1] == '<' {
						depth++
						i += 2
						continue
					}
					if i+1 < len(out) && out[i] == '>' && out[i+1] == '>' {
						depth--
						i += 2
						if depth == 0 {
							dictEnd = i - 2
							break
						}
						continue
					}
					i++
				}
				if dictEnd >= 0 {
					dictBytes := out[dictStart : dictEnd+2]
					// Remove existing NeedAppearances if present
					needAppRe := regexp.MustCompile(`\s*/NeedAppearances\s+(?:true|false)`)
					if needAppRe.Match(dictBytes) {
						out = needAppRe.ReplaceAll(out, []byte(""))
						dictEnd -= len(needAppRe.Find(dictBytes))
					}
					// Add NeedAppearances false
					insertion := []byte(" /NeedAppearances false")
					out = append(out[:dictEnd], append(insertion, out[dictEnd:]...)...)
				}
			}
		}
	}

	for _, job := range jobs {
		// Use original text for display (no extra escaping)
		displayText := job.text
		// Only escape for the PDF stream syntax
		streamText := escapePDFString(job.text)

		// Calculate text positioning based on alignment
		var tx float64
		textWidth := float64(len(displayText)) * job.fontSize * 0.6
		switch job.q {
		case 1: // Center
			tx = (job.width - textWidth) / 2
		case 2: // Right
			tx = job.width - textWidth - 3
		default: // Left
			tx = 3
		}
		if tx < 3 {
			tx = 3
		}

		// Center text vertically in field with better calculation
		y := (job.height-job.fontSize)/2 + 2
		if y < 3 {
			y = 3
		}

		// Determine font base name for resources
		baseFontName := "Helvetica"
		switch job.fontResourceName {
		case "TiRo", "Times-Roman", "Times":
			baseFontName = "Times-Roman"
		case "TiBo", "Times-Bold":
			baseFontName = "Times-Bold"
		case "TiIt", "Times-Italic":
			baseFontName = "Times-Italic"
		case "TiBI", "Times-BoldItalic":
			baseFontName = "Times-BoldItalic"
		case "Cour", "Courier":
			baseFontName = "Courier"
		case "CoBo", "Courier-Bold":
			baseFontName = "Courier-Bold"
		case "CoOb", "Courier-Oblique":
			baseFontName = "Courier-Oblique"
		case "CoBO", "Courier-BoldOblique":
			baseFontName = "Courier-BoldOblique"
		case "Helv", "Helvetica":
			baseFontName = "Helvetica"
		case "HeBo", "Helvetica-Bold":
			baseFontName = "Helvetica-Bold"
		case "HeOb", "Helvetica-Oblique":
			baseFontName = "Helvetica-Oblique"
		case "HeBO", "Helvetica-BoldOblique":
			baseFontName = "Helvetica-BoldOblique"
		default:
			baseFontName = "Helvetica"
		}

		// Create enhanced appearance stream with proper PDF operators for Adobe compatibility
		streamBody := fmt.Sprintf("q\n"+
			"0 0 %.2f %.2f re W n\n"+ // Clipping rectangle
			"1 1 1 rg\n"+ // White background
			"0 0 %.2f %.2f re f\n"+ // Fill background
			"BT\n"+
			"0 0 0 rg\n"+ // Black text color
			"/%s %.2f Tf\n"+
			"%.2f %.2f Td\n"+
			"(%s) Tj\n"+
			"ET\n"+
			"Q\n", job.width, job.height, job.width, job.height, job.fontResourceName, job.fontSize, tx, y, streamText)

		// Create XObject appearance stream with enhanced font resources and proper structure
		apObj := fmt.Sprintf("%d 0 obj\n"+
			"<< /Type /XObject\n"+
			"   /Subtype /Form\n"+
			"   /FormType 1\n"+
			"   /BBox [0 0 %.2f %.2f]\n"+
			"   /Matrix [1 0 0 1 0 0]\n"+
			"   /Resources << /Font << /%s << /Type /Font /Subtype /Type1 /BaseFont /%s /Encoding /WinAnsiEncoding >> >> >>\n"+
			"   /Length %d\n"+
			">>\n"+
			"stream\n"+
			"%s"+
			"endstream\n"+
			"endobj\n",
			job.apObjNum, job.width, job.height, job.fontResourceName, baseFontName, len(streamBody), streamBody)

		out = append(out, []byte(apObj)...)
	}

	// rebuild xref
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
	xrefBuf := bytes.NewBuffer(nil)
	xrefBuf.WriteString(fmt.Sprintf("xref\n0 %d\n", maxObj+1))
	xrefBuf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			xrefBuf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
		} else {
			xrefBuf.WriteString("0000000000 00000 f \n")
		}
	}

	// find Root from original pdf
	root := 1
	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s0\sR`)
	if rm := rootRe.FindSubmatch(pdfBytes); len(rm) > 1 {
		if r, err := strconv.Atoi(string(rm[1])); err == nil {
			root = r
		}
	}

	trailer := fmt.Sprintf("trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%EOF\n", maxObj+1, root, xrefStart)
	out = append(out, xrefBuf.Bytes()...)
	out = append(out, []byte(trailer)...)

	// After creating appearance streams and rebuilding xref, call FlattenPDFBytes
	flat, err := FlattenPDFBytes(out)
	if err != nil {
		return nil, err
	}
	return flat, nil
}

// FlattenPDFBytes flattens form fields from the provided PDF bytes and returns flattened PDF bytes.
// This is a local stub so callers in this package can invoke flattening. Replace implementation
// to call the shared flattener when available.
func FlattenPDFBytes(pdfBytes []byte) ([]byte, error) {
	// TODO: call the canonical flattener (e.g., scripts.FlattenPDFBytes) once moved into an importable package.
	// For now, return the input bytes unchanged.
	return scripts.FlattenPDFBytes(pdfBytes)
}
