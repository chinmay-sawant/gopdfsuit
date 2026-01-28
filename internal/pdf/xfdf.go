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
						if _, err := fmt.Sscanf(string(fm[1]), "%d", &first); err != nil {
							first = 0
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
	widgetRe := regexp.MustCompile(`(?s)<<.*?/Subtype\s*/Widget.*?>>`)
	nameRe := regexp.MustCompile(`/T\s*\((.*?)\)`)
	widgetMatches := widgetRe.FindAllIndex(out, -1)

	discoveredWidgets := make(map[string][]job)

	for _, match := range widgetMatches {
		dictStart, dictEnd := match[0], match[1]
		dictBytes := out[dictStart:dictEnd]
		nameMatch := nameRe.FindSubmatch(dictBytes)
		if nameMatch == nil {
			continue
		}
		fieldName := string(nameMatch[1])

		if _, ok := fields[fieldName]; !ok {
			continue // Skip if this field is not in our input data
		}

		currentVal := fields[fieldName]
		newJob := job{field: fieldName, val: currentVal, dictStart: dictStart, dictEnd: dictEnd}

		if bytes.Contains(dictBytes, []byte("/FT /Btn")) {
			if bytes.Contains(dictBytes, []byte("/Parent")) {
				newJob.fieldType = typeRadio
				apDictRe := regexp.MustCompile(`/AP\s*<<.*?/N\s*<<\s*/\s*([A-Za-z0-9_]+)\s*`)
				if apMatch := apDictRe.FindSubmatch(dictBytes); apMatch != nil {
					newJob.radioExportValue = string(apMatch[1])
				}
			} else {
				newJob.fieldType = typeButton
			}
		} else if bytes.Contains(dictBytes, []byte("/FT /Tx")) {
			newJob.fieldType = typeText
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
			newJob.width, newJob.height = urx-llx, ury-lly

			qRe := regexp.MustCompile(`/Q\s*(\d)`)
			if qMatch := qRe.FindSubmatch(dictBytes); qMatch != nil {
				newJob.q, _ = strconv.Atoi(string(qMatch[1]))
			}
			newJob.fontSize, newJob.fontResourceName = 12.0, "Helv" // Default to Helv
			daRe := regexp.MustCompile(`/DA\s*\((.*?)\)`)
			if daMatch := daRe.FindSubmatch(dictBytes); daMatch != nil {
				tfRe := regexp.MustCompile(`/([\w.-]+)\s+([\d.]+)\s+Tf`)
				if tfMatch := tfRe.FindStringSubmatch(string(daMatch[1])); len(tfMatch) > 2 {
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

	// --- PASS 2: MODIFICATION (WRITE-ONLY, IN REVERSE) ---
	sort.Slice(allJobs, func(i, j int) bool { return allJobs[i].dictStart > allJobs[j].dictStart })

	radioGroups := make(map[string]string)
	for _, job := range allJobs {
		if job.fieldType == typeRadio {
			radioGroups[job.field] = job.val
		}
	}
	for fieldName, value := range radioGroups {
		re := regexp.MustCompile(fmt.Sprintf(`(?s)(<<.*?/T\s*\(\s*%s\s*\).*?/Kids.*?>>)`, regexp.QuoteMeta(fieldName)))
		match := re.FindIndex(out)
		if match == nil || bytes.Contains(out[match[0]:match[1]], []byte("/Subtype /Widget")) {
			continue
		}
		dictStart, dictEnd := match[0], match[1]
		dictBytes := out[dictStart:dictEnd]
		newV := []byte("/V /" + value)
		vRe := regexp.MustCompile(`/V\s*\(?.*?\)?`)
		var newDictBytes []byte
		if vRe.Match(dictBytes) {
			newDictBytes = vRe.ReplaceAll(dictBytes, newV)
		} else {
			newDictBytes = bytes.Replace(dictBytes, []byte(">>"), append([]byte(" "), append(newV, []byte(">>")...)...), 1)
		}
		out = append(out[:dictStart], append(newDictBytes, out[dictEnd:]...)...)
	}

	for _, job := range allJobs {
		dictBytes := out[job.dictStart:job.dictEnd]
		var newDictBytes []byte
		switch job.fieldType {
		case typeText:
			esc := escapePDFString(job.val)
			newV := []byte(fmt.Sprintf("/V (%s)", esc))
			vRe := regexp.MustCompile(`/V\s*\(.*?\)`)
			if vRe.Match(dictBytes) {
				newDictBytes = vRe.ReplaceAll(dictBytes, newV)
			} else {
				newDictBytes = bytes.Replace(dictBytes, []byte(">>"), append([]byte(" "), append(newV, []byte(">>")...)...), 1)
			}
			apRe := regexp.MustCompile(`\s*/AP\s*<<.*?>>`)
			newDictBytes = apRe.ReplaceAll(newDictBytes, []byte(" "))
		case typeButton, typeRadio:
			newState := "/Off"
			if job.fieldType == typeButton && (strings.ToLower(job.val) == "yes" || strings.ToLower(job.val) == "on") {
				apDictRe := regexp.MustCompile(`/AP\s*<<.*?/N\s*<<\s*/\s*([A-Za-z0-9_]+)\s*`)
				if apMatch := apDictRe.FindSubmatch(dictBytes); apMatch != nil {
					newState = "/" + string(apMatch[1])
				} else {
					newState = "/Yes"
				}
			} else if job.fieldType == typeRadio && job.radioExportValue == job.val {
				newState = "/" + job.radioExportValue
			}
			newAS := []byte("/AS " + newState)
			asRe := regexp.MustCompile(`/AS\s*/\w+`)
			if asRe.Match(dictBytes) {
				newDictBytes = asRe.ReplaceAll(dictBytes, newAS)
			} else {
				newDictBytes = bytes.Replace(dictBytes, []byte(">>"), append([]byte(" "), append(newAS, []byte(">>")...)...), 1)
			}
		}
		if newDictBytes != nil {
			out = append(out[:job.dictStart], append(newDictBytes, out[job.dictEnd:]...)...)
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
		re := regexp.MustCompile(fmt.Sprintf(`(?s)(<<.*?/Subtype\s*/Widget.*?/T\s*\(\s*%s\s*\).*?>>)`, regexp.QuoteMeta(job.field)))
		match := re.FindIndex(out)
		if match == nil {
			continue
		}
		dictEnd := match[1]
		apRef := []byte(fmt.Sprintf(" /AP<</N %d 0 R>>", job.apObjNum))
		out = append(out[:dictEnd-2], append(apRef, out[dictEnd-2:]...)...)
	}

	needAppRe := regexp.MustCompile(`/NeedAppearances\s+(true|false)`)
	if needAppRe.Match(out) {
		out = needAppRe.ReplaceAll(out, []byte("/NeedAppearances false"))
	} else {
		acroFormRe := regexp.MustCompile(`(/AcroForm\s*<<)`)
		if loc := acroFormRe.FindIndex(out); loc != nil {
			insertPos := loc[1]
			insertContent := []byte(" /NeedAppearances false ")
			newOut := make([]byte, 0, len(out)+len(insertContent))
			newOut = append(newOut, out[:insertPos]...)
			newOut = append(newOut, insertContent...)
			newOut = append(newOut, out[insertPos:]...)
			out = newOut
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
		streamBody := fmt.Sprintf("q\nBT\n/F1 %.2f Tf\n0 g\n%.2f %.2f Td\n(%s) Tj\nET\nQ",
			job.fontSize, tx, y, streamText)

		fontDescObj := fmt.Sprintf("\n%d 0 obj\n<</Type/FontDescriptor/FontName/%s/Flags 32/FontBBox[-558 -225 1000 931]/ItalicAngle 0/Ascent 905/Descent -212/CapHeight 905/StemV 88>>\nendobj\n",
			job.fontDescObjNum, job.fontResourceName)
		out = append(out, []byte(fontDescObj)...)

		// --- START OF CHANGES ---

		// Build the widths array string from the constant.
		var widthsBuf strings.Builder
		widthsBuf.WriteString("[")
		for i, w := range helveticaWidths {
			widthsBuf.WriteString(strconv.Itoa(w))
			if i < len(helveticaWidths)-1 {
				widthsBuf.WriteString(" ")
			}
		}
		widthsBuf.WriteString("]")
		widthsStr := widthsBuf.String()

		// Update the Font object to include FirstChar, LastChar, and the Widths array.
		// Using full WinAnsiEncoding range (32-255) for PDF 2.0 compliance
		fontObj := fmt.Sprintf("\n%d 0 obj\n<</Type/Font/Subtype/Type1/BaseFont/%s/Encoding/WinAnsiEncoding/FirstChar 32/LastChar 255/Widths %s/FontDescriptor %d 0 R>>\nendobj\n",
			job.fontObjNum, job.fontResourceName, widthsStr, job.fontDescObjNum)

		// --- END OF CHANGES ---

		out = append(out, []byte(fontObj)...)

		var compBuf bytes.Buffer
		zw, _ := zlib.NewWriterLevel(&compBuf, zlib.BestCompression)
		if _, err := zw.Write([]byte(streamBody)); err != nil {
			return nil, fmt.Errorf("compression write failed: %w", err)
		}
		if err := zw.Close(); err != nil {
			return nil, fmt.Errorf("compression close failed: %w", err)
		}
		comp := compBuf.Bytes()
		apObj := fmt.Sprintf("\n%d 0 obj\n<</Type/XObject/Subtype/Form/FormType 1/BBox[0 0 %.2f %.2f]/Resources<</Font<</F1 %d 0 R>>/ProcSet[/PDF/Text]>>/Filter/FlateDecode/Length %d>>\nstream\n%s\nendstream\nendobj\n",
			job.apObjNum, job.width, job.height, job.fontObjNum, len(comp), string(comp))
		out = append(out, []byte(apObj)...)
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
	fmt.Fprintf(&xrefBuf, "xref\n0 %d\n", maxObj+1)
	xrefBuf.WriteString("0000000000 65535 f \r\n")
	for i := 1; i <= maxObj; i++ {
		if off, ok := offsets[i]; ok {
			fmt.Fprintf(&xrefBuf, "%010d 00000 n \r\n", off)
		} else {
			xrefBuf.WriteString("0000000000 65535 f \r\n")
		}
	}
	root := 1
	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s+0\s+R`)
	if rm := rootRe.FindSubmatch(pdfBytes); len(rm) > 1 {
		if r, err := strconv.Atoi(string(rm[1])); err == nil {
			root = r
		}
	}
	trailer := fmt.Sprintf("trailer\n<</Size %d/Root %d 0 R>>\nstartxref\n%d\n%%%%EOF\n", maxObj+1, root, xrefStart)
	out = append(out, xrefBuf.Bytes()...)
	out = append(out, []byte(trailer)...)
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
