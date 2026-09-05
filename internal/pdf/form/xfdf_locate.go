// Package form provides functionality for parsing XFDF and filling PDF forms.
package form

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/pdfobj"
)

const maxFieldTreeDepth = 1000

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

// tryZlibDecompress attempts to decompress zlib data.
func tryZlibDecompress(b []byte) ([]byte, error) {
	return pdfobj.InflateZlib(b)
}

// tryFlateDecompress attempts to decompress raw flate data.
func tryFlateDecompress(b []byte) ([]byte, error) {
	return pdfobj.InflateFlate(b)
}

// extractTokenGroups looks for /V or /AS tokens near a position
func extractTokenGroups(content []byte, pos int) (string, string) {
	limit := min(pos+800, len(content))
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

// findRootRef looks for /Root n m R in the PDF bytes.
func findRootRef(data []byte) (string, bool) {
	n, g, ok := pdfobj.FindRootRef(data)
	if !ok {
		return "", false
	}
	return strconv.Itoa(n) + " " + strconv.Itoa(g), true
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
func traverseField(ref string, objMap map[string][]byte, parentPrefix string, out map[string]string, active map[string]bool, depth int) error {
	if depth > maxFieldTreeDepth {
		return fmt.Errorf("field tree exceeds maximum depth %d", maxFieldTreeDepth)
	}
	if active[ref] {
		return fmt.Errorf("field tree cycle at object %s", ref)
	}
	body, ok := objMap[ref]
	if !ok {
		return nil
	}
	active[ref] = true
	defer delete(active, ref)

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
			if err := traverseField(kidRef, objMap, fullName, out, active, depth+1); err != nil {
				return err
			}
		}
	}
	if m := reSingleKids.FindSubmatch(body); m != nil {
		kidRef := string(m[1]) + " " + string(m[2])
		if err := traverseField(kidRef, objMap, fullName, out, active, depth+1); err != nil {
			return err
		}
	}
	return nil
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

	// Build the field object map (object-stream expansion included).
	objMap := buildFieldObjectMap(pdfBytes)

	if len(objMap) == 0 {
		// Fall back to naive scan
		return detectFormFieldsNaive(pdfBytes)
	}

	// Try the structured locate seam first.
	structured, err := locateFields(objMap, pdfBytes)
	if err != nil {
		return nil, err
	}
	if len(structured) > 0 {
		return structured, nil
	}

	// Fall back to naive detection
	return detectFormFieldsNaive(pdfBytes)
}

// buildFieldObjectMap builds the "num gen" -> body object map for field
// location: top-level objects, object-stream expansion, stream
// decompression, and xref-stream augmentation via the pdfobj read seam.
func buildFieldObjectMap(pdfBytes []byte) map[string][]byte {
	objMap := make(map[string][]byte)
	for _, m := range reObjStream.FindAllSubmatch(pdfBytes, -1) {
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

	// Attempt to parse XRef streams to augment object map.
	parseXRefStreams(pdfBytes, objMap)

	return objMap
}

// LocateFields is the unit-testable field-locate seam: it walks an
// already-built object map from the catalog's /AcroForm /Fields and
// resolves field names and values. It needs no full PDF, only the map
// plus the raw bytes for the /Root fallback lookup.
func LocateFields(objMap map[string][]byte, pdfBytes []byte) map[string]string {
	structured, _ := locateFields(objMap, pdfBytes)
	return structured
}

func locateFields(objMap map[string][]byte, pdfBytes []byte) (map[string]string, error) {
	structured := make(map[string]string)
	active := make(map[string]bool)
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
							if err := traverseField(fref, objMap, "", structured, active, 0); err != nil {
								return nil, err
							}
						}
					} else {
						singleFields := regexp.MustCompile(`/Fields\s+(\d+)\s+(\d+)\s+R`)
						if sm := singleFields.FindSubmatch(afBody); sm != nil {
							fref := string(sm[1]) + " " + string(sm[2])
							if err := traverseField(fref, objMap, "", structured, active, 0); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
	}

	return structured, nil
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
