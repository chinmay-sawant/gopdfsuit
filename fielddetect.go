package main

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// fielddetect.go
// Simple, self-contained heuristic detector for PDF AcroForm fields.
// It does NOT fully parse PDF objects or resolve indirect references.
// Instead it scans the raw PDF bytes for /T (field name) tokens and
// looks nearby for /V or /AS tokens to heuristically extract values.

func decodeHexString(s string) string {
	// remove whitespace
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	if len(s)%2 == 1 {
		// odd length, pad with 0
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return "<invalid hex>"
	}
	// PDF text is usually ASCII or Latin-1; we'll convert bytes to string directly
	return string(b)
}

func extractTokenGroups(content []byte, pos int) (string, string) {
	// Look ahead up to a limit for /V or /AS tokens
	limit := pos + 800
	if limit > len(content) {
		limit = len(content)
	}
	window := content[pos:limit]

	// regex matching literal string (parens), hex <...>, or name /Name
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fielddetect <file.pdf>")
		fmt.Println("Default: sampledata/patientreg/patientreg.pdf")
		os.Exit(1)
	}

	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Opened PDF %s (%d bytes)\n", path, len(data))

	if off, ok := findStartXref(data); ok {
		fmt.Printf("startxref at %d\n", off)
	}
	if trailerHasEncrypt(data) {
		fmt.Println("Warning: PDF appears to be encrypted or contain /Encrypt entries — this tool cannot decrypt files yet.")
	}

	// Build map of indirect objects: "<obj> <gen>" -> body
	// Use (?s) to allow dot to match newlines
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	objMatches := objRe.FindAllSubmatch(data, -1)
	if len(objMatches) == 0 {
		// Fall back to naive scan
		tRe := regexp.MustCompile(`/T\s*(\(([^)]*)\)|<([0-9A-Fa-f\s]+)>|/([A-Za-z0-9#]+))`)
		matches := tRe.FindAllSubmatchIndex(data, -1)
		if len(matches) == 0 {
			fmt.Println("No /T field tokens found in PDF.")
			os.Exit(0)
		}
		fmt.Printf("Detected %d candidate fields (flat scan):\n", len(matches))
		seen := make(map[string]bool)
		for _, mi := range matches {
			var name string
			if mi[4] != -1 && mi[5] != -1 {
				name = string(data[mi[4]:mi[5]])
			} else if mi[6] != -1 && mi[7] != -1 {
				name = decodeHexString(string(data[mi[6]:mi[7]]))
			} else if mi[8] != -1 && mi[9] != -1 {
				name = string(data[mi[8]:mi[9]])
			} else {
				continue
			}
			name = strings.TrimSpace(name)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			endPos := mi[1]
			tType, val := extractTokenGroups(data, endPos)
			if tType == "" {
				fmt.Printf("- Key: %s  Value: <not found nearby>\n", name)
			} else {
				fmt.Printf("- Key: %s  %s: %s\n", name, tType, val)
			}
		}
		os.Exit(0)
	}

	objMap := make(map[string][]byte)
	for _, m := range objMatches {
		// m[1]=objnum, m[2]=gen, m[3]=body
		key := string(m[1]) + " " + string(m[2])
		body := m[3]
		// If this object is an ObjStm, try to extract its embedded objects
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
		// attempt to locate any stream...endstream sections and decompress when Flate
		streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
		newBody := body
		changed := false
		for _, sm := range streamRe.FindAllSubmatchIndex(body, -1) {
			// sm gives start/end indices for full match and capture group
			sStart := sm[2]
			sEnd := sm[3]
			if sStart < 0 || sEnd < 0 || sEnd <= sStart {
				continue
			}
			streamBytes := body[sStart:sEnd]
			// try zlib then raw flate
			var dec []byte
			if d, err := tryZlibDecompress(streamBytes); err == nil {
				dec = d
			} else if d, err := tryFlateDecompress(streamBytes); err == nil {
				dec = d
			}
			if dec != nil {
				// replace the stream bytes with the decompressed content (as plain bytes)
				// build newBody using bytes.Buffer
				var buf bytes.Buffer
				buf.Write(newBody[:sm[0]])
				buf.Write(dec)
				buf.Write(newBody[sm[1]:])
				newBody = buf.Bytes()
				changed = true
				// continue but avoid overlapping indexes (we restart search on newBody)
				break
			}
		}
		if changed {
			// if we changed, re-run to decompress any further streams
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
		}
		objMap[key] = newBody
	}

	fmt.Printf("Parsed %d indirect objects from PDF.\n", len(objMap))

	// Attempt to parse XRef streams to augment object map
	parseXRefStreams(data, objMap)

	// Try to locate the AcroForm via Root->/AcroForm and traverse fields
	structured := make(map[string]string)
	if rootRef, ok := findRootRef(data); ok {
		if rootBody, ok2 := objMap[rootRef]; ok2 {
			if acroRef, ok3 := getAcroFormRef(rootBody, data); ok3 {
				// acroRef should be like "n m"
				if afBody, ok4 := objMap[acroRef]; ok4 {
					// find /Fields array
					fieldsRe := regexp.MustCompile(`/Fields\s*\[(.*?)\]`)
					if fm := fieldsRe.FindSubmatch(afBody); fm != nil {
						inner := fm[1]
						refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
						for _, r := range refRe.FindAllSubmatch(inner, -1) {
							fref := string(r[1]) + " " + string(r[2])
							traverseField(fref, objMap, "", structured)
						}
					} else {
						// maybe Fields is a single ref
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
		fmt.Println("STRUCTURED FIELDS:")
		for k, v := range structured {
			if v == "" {
				fmt.Printf("- %s: <empty>\n", k)
			} else {
				fmt.Printf("- %s: %s\n", k, v)
			}
		}
	}
}

// bytesIndex is a tiny helper to find a subsequence in a []byte without
// importing bytes package repeatedly; keeps code simple.
func bytesIndex(b, sub []byte) int {
	return strings.Index(string(b), string(sub))
}

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

func tryFlateDecompress(b []byte) ([]byte, error) {
	// flate.NewReader expects raw DEFLATE stream
	r := flate.NewReader(bytes.NewReader(b))
	defer r.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

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

func readUint(b []byte) uint64 {
	var v uint64
	for _, c := range b {
		v = (v << 8) | uint64(byte(c))
	}
	return v
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
		// Simpler: parse sequentially across dec using W widths
		w0, w1, w2 := W[0], W[1], W[2]
		total := w0 + w1 + w2
		idx := 0
		// determine starting objnum from Index first pair
		// not tracking starting object number precisely here
		for pos := 0; pos+total <= len(dec); pos += total {
			f1 := int(readUint(dec[pos : pos+w0]))
			f2 := int(readUint(dec[pos+w0 : pos+w0+w1]))
			f3 := int(readUint(dec[pos+w0+w1 : pos+total]))
			idx++
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
					// naive split by whitespace using offsets isn't easy; skip detailed extraction here
					_ = index
					_ = stm
				}
			}
		}
	}
}

// findRootRef looks for /Root n m R in the whole PDF bytes
func findRootRef(data []byte) (string, bool) {
	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	if m := rootRe.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// getAcroFormRef finds /AcroForm n m R in given object body or in the whole file
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

// ...existing code...

// traverseField resolves a field object (by ref key like "123 0") and extracts full field names and values.
// parentPrefix is preprended with a dot if non-empty when combining parent and child names.
func traverseField(ref string, objMap map[string][]byte, parentPrefix string, out map[string]string) {
	body, ok := objMap[ref]
	if !ok {
		return
	}

	// find /T
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
		// attempt to get /V or /AS in this object
		endPos := m[1]
		tType, val := extractTokenGroups(body, endPos)
		if tType != "" {
			tv = val
		} else {
			// try resolving indirect /V refs
			if rv := resolveValueRef(body, objMap); rv != "" {
				tv = rv
			}
			// try widget lookup (AS/AP) for checkboxes/radios
			if tv == "" {
				if asn, ok := findWidgetAnnotationsForName(name, objMap); ok {
					tv = asn
				}
			}
		}
	}

	// build full name
	fullName := name
	if parentPrefix != "" && name != "" {
		fullName = parentPrefix + "." + name
	} else if parentPrefix != "" && name == "" {
		fullName = parentPrefix
	}

	// if we have a name, record value (may be empty)
	if fullName != "" {
		if tv == "" {
			out[fullName] = ""
		} else {
			out[fullName] = tv
		}
	}

	// If value wasn't found, check /Kids and recurse
	kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	if m := kidsRe.FindSubmatch(body); m != nil {
		inner := m[1]
		for _, r := range refRe.FindAllSubmatch(inner, -1) {
			kidRef := string(r[1]) + " " + string(r[2])
			traverseField(kidRef, objMap, fullName, out)
		}
	}
	// Also check /Kids as single ref (rare)
	singleKidsRe := regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	if m := singleKidsRe.FindSubmatch(body); m != nil {
		kidRef := string(m[1]) + " " + string(m[2])
		traverseField(kidRef, objMap, fullName, out)
	}

	// If no /T but there are kids, propagate prefix
	// Also check /Annots on pages: these can contain widget refs; we'll pick those up via candidate scanning elsewhere
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

// resolveValueRef attempts to resolve a /V value which can be an indirect ref like '123 0 R'
// or a literal inside the given object body. If a ref is found, follow to referenced object.
func resolveValueRef(body []byte, objMap map[string][]byte) string {
	// recursive resolver with depth limit
	var resolve func(b []byte, depth int) string
	resolve = func(b []byte, depth int) string {
		if depth > 6 {
			return ""
		}
		// direct literal nearby
		if tType, v := extractTokenGroups(b, 0); tType != "" && v != "" {
			return v
		}
		// try to extract any literal from the bytes (including stream contents)
		if s := extractStringFromBytes(b); s != "" {
			return s
		}
		// check for /V n m R (indirect)
		refRe := regexp.MustCompile(`/V\s*(\d+)\s+(\d+)\s+R`)
		if m := refRe.FindSubmatch(b); m != nil {
			ref := string(m[1]) + " " + string(m[2])
			if rb, ok := objMap[ref]; ok {
				// if referenced object has a stream, decompress and search inside
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
				// try recursion on the referenced object's bytes
				if s := resolve(rb, depth+1); s != "" {
					return s
				}
				// last resort: try to find literal tokens anywhere in referenced bytes
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

// extractStringFromBytes looks for common PDF literal representations in arbitrary bytes:
// - parentheses literal ( ... )
// - hex literal <A1B2>
// - name /Name
func extractStringFromBytes(b []byte) string {
	// search for literal paren string
	parenRe := regexp.MustCompile(`\(([^)]{1,200})\)`)
	if m := parenRe.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	// hex string
	hexRe := regexp.MustCompile(`<([0-9A-Fa-f\s]{2,400})>`)
	if m := hexRe.FindSubmatch(b); m != nil {
		return decodeHexString(string(m[1]))
	}
	// name
	nameRe := regexp.MustCompile(`/([A-Za-z0-9_+-]{1,200})`)
	if m := nameRe.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	return ""
}

// findWidgetAnnotationsForName searches objMap for widget annotation objects that contain the field name
// and attempts to extract their /AS or appearance /AP normal state name.
func findWidgetAnnotationsForName(name string, objMap map[string][]byte) (string, bool) {
	// look for literal (name) occurrences
	needle := []byte("(" + name + ")")
	for k, body := range objMap {
		if bytesIndex(body, []byte(`/Subtype/Widget`)) < 0 {
			continue
		}
		if bytesIndex(body, needle) < 0 && bytesIndex(body, []byte(`/T`)) < 0 {
			continue
		}
		// If T matches name in this widget, check AS
		if bytesIndex(body, needle) >= 0 {
			// check for /AS
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
			// check /AP /N dict keys for a named appearance (like /Yes)
			apRe := regexp.MustCompile(`/AP\s*<<(.*?)>>`)
			if am := apRe.FindSubmatch(body); am != nil {
				nRe := regexp.MustCompile(`/N\s*<<(.*?)>>`)
				if nm := nRe.FindSubmatch(am[1]); nm != nil {
					// nm[1] contains inner dict; extract any name keys
					keyRe := regexp.MustCompile(`/([A-Za-z0-9_+-]+)\s*(?:/|stream|<<|\()`)
					if kr := keyRe.FindSubmatch(nm[1]); kr != nil {
						return string(kr[1]), true
					}
				}
				// /AP /N could be a name or stream directly
				nNameRe := regexp.MustCompile(`/N\s*/([A-Za-z0-9_+-]+)`)
				if nn := nNameRe.FindSubmatch(am[1]); nn != nil {
					return string(nn[1]), true
				}
			}
			// fallback: report widget key
			return k, true
		}
	}
	return "", false
}
