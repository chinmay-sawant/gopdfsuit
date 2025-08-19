package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: diag <pdf-file>")
		os.Exit(1)
	}
	path := os.Args[1]
	targets := map[string]struct{}{}
	if len(os.Args) > 2 {
		for _, t := range strings.Split(os.Args[2], ",") {
			targets[strings.TrimSpace(t)] = struct{}{}
		}
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	fmt.Printf("Analyzing PDF: %s\n\n", path)

	// Find all field names via /T (name)
	reT := regexp.MustCompile(`/T\s*\(([^\)]*)\)`)
	matches := reT.FindAllSubmatchIndex(b, -1)
	if len(matches) == 0 {
		fmt.Println("No /T entries found in PDF.")
		return
	}

	// Helper regexes
	reFT := regexp.MustCompile(`/FT\s*/(\w+)`)
	reVStr := regexp.MustCompile(`/V\s*\(([^\)]*)\)`)
	reVName := regexp.MustCompile(`/V\s*/(\w+)`)
	reAS := regexp.MustCompile(`/AS\s*/(\w+)`)
	reAP := regexp.MustCompile(`/AP\s*<<([^>]*)>>`)
	reRect := regexp.MustCompile(`/Rect\s*\[([^\]]*)\]`)
	reDA := regexp.MustCompile(`/DA\s*\(([^\)]*)\)`)
	reFf := regexp.MustCompile(`/Ff\s*(\d+)`)
	reF := regexp.MustCompile(`/F\s*(\d+)`)

	for i, m := range matches {
		name := string(b[m[2]:m[3]])
		pos := m[0]
		// context window
		ctxStart := pos - 4096
		if ctxStart < 0 {
			ctxStart = 0
		}
		ctxEnd := pos + 4096
		if ctxEnd > len(b) {
			ctxEnd = len(b)
		}
		ctx := b[ctxStart:ctxEnd]

		// find nearest dict start before pos
		relStart := bytes.LastIndex(ctx[:pos-ctxStart+1], []byte("<<"))
		dictStart := -1
		dictEnd := -1
		if relStart >= 0 {
			dictStart = ctxStart + relStart
			// find matching >>
			depth := 0
			j := dictStart
			for j < ctxEnd-1 {
				if j+1 < len(b) && b[j] == '<' && b[j+1] == '<' {
					depth++
					j += 2
					continue
				}
				if j+1 < len(b) && b[j] == '>' && b[j+1] == '>' {
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
		}

		var dict []byte
		if dictStart >= 0 && dictEnd >= 0 && dictEnd > dictStart {
			dict = b[dictStart : dictEnd+2]
		} else {
			// fallback: use window
			dict = ctx
		}

		fmt.Printf("%d) Field name: '%s' (pos=%d)\n", i+1, name, pos)
		if dictStart >= 0 {
			fmt.Printf("   widget dict range: %d - %d\n", dictStart, dictEnd)
		} else {
			fmt.Printf("   widget dict: not found in 4KB window; showing nearby context\n")
		}

		// extract properties
		if ft := reFT.FindSubmatch(dict); ft != nil {
			fmt.Printf("   /FT: %s\n", string(ft[1]))
		}
		if rect := reRect.FindSubmatch(dict); rect != nil {
			coords := strings.TrimSpace(string(rect[1]))
			parts := strings.Fields(coords)
			if len(parts) >= 4 {
				fmt.Printf("   /Rect: %s -> llx=%s lly=%s urx=%s ury=%s\n", coords, parts[0], parts[1], parts[2], parts[3])
			} else {
				fmt.Printf("   /Rect: %s\n", coords)
			}
		}
		if v := reVStr.FindSubmatch(dict); v != nil {
			fmt.Printf("   /V (string): %s\n", string(v[1]))
		} else if vname := reVName.FindSubmatch(dict); vname != nil {
			fmt.Printf("   /V (name): /%s\n", string(vname[1]))
		} else {
			fmt.Printf("   /V: not present in the widget dict\n")
		}
		if as := reAS.FindSubmatch(dict); as != nil {
			fmt.Printf("   /AS: /%s\n", string(as[1]))
		}
		if ap := reAP.FindSubmatch(dict); ap != nil {
			inner := string(ap[1])
			// find numeric refs inside
			reRef := regexp.MustCompile(`(\d+)\s+0\s+R`)
			refs := reRef.FindAllStringSubmatch(inner, -1)
			if len(refs) > 0 {
				fmt.Printf("   /AP refs:")
				for _, r := range refs {
					fmt.Printf(" %s 0 R", r[1])
				}
				fmt.Printf("\n")
			} else {
				fmt.Printf("   /AP present but no numeric object refs detected\n")
			}
		} else {
			// check if AP exists elsewhere near dict by searching for "/AP"
			if bytes.Contains(dict, []byte("/AP")) {
				fmt.Printf("   /AP present (raw)\n")
			} else {
				fmt.Printf("   /AP: not present in this dict\n")
			}
		}
		if da := reDA.FindSubmatch(dict); da != nil {
			fmt.Printf("   /DA: %s\n", string(da[1]))
		}
		if ff := reFf.FindSubmatch(dict); ff != nil {
			if n, err := strconv.Atoi(string(ff[1])); err == nil {
				fmt.Printf("   /Ff: %d\n", n)
			}
		}
		if f := reF.FindSubmatch(dict); f != nil {
			if n, err := strconv.Atoi(string(f[1])); err == nil {
				fmt.Printf("   /F (ann flags): %d\n", n)
			}
		}

		// check appearance stream objects referenced by /AP
		apObjRe := regexp.MustCompile(`/AP\s*<<[^>]*?(\d+)\s+0\s+R`)
		if a := apObjRe.FindSubmatch(dict); a != nil {
			objNum := string(a[1])
			// search for obj header
			hdr := []byte(fmt.Sprintf("%s 0 obj", objNum))
			if idx := bytes.Index(b, hdr); idx >= 0 {
				fmt.Printf("   AP object %s found at offset %d\n", objNum, idx)
			} else {
				fmt.Printf("   AP object %s referenced but not found in file bytes\n", objNum)
			}
		}

		// quick heuristics
		isButton := bytes.Contains(dict, []byte("/FT /Btn")) || bytes.Contains(dict, []byte("/AS "))
		if isButton {
			fmt.Printf("   heuristic: this looks like a button/checkbox/radio field\n")
		}

		// If this field is in targets, print the containing object for closer inspection
		if _, ok := targets[name]; ok {
			// search backwards for an obj header like '123 0 obj' before dictStart
			objHeaderRe := regexp.MustCompile(`(\d+)\s+0\s+obj`)
			searchStart := dictStart
			if searchStart < 0 {
				searchStart = pos
			}
			if searchStart < 2000 {
				searchStart = 0
			} else {
				searchStart = searchStart - 2000
			}
			searchBytes := b[searchStart : dictStart+1]
			objMatches := objHeaderRe.FindAllSubmatchIndex(searchBytes, -1)
			if len(objMatches) > 0 {
				last := objMatches[len(objMatches)-1]
				objNum := string(searchBytes[last[2]:last[3]])
				objPos := searchStart + last[0]
				// find end of object (endobj)
				endPos := bytes.Index(b[objPos:], []byte("endobj"))
				if endPos >= 0 {
					objBytes := b[objPos : objPos+endPos+6]
					fmt.Printf("   containing object header: %s found at %d; printing first 400 bytes:\n", objNum, objPos)
					snippet := objBytes
					if len(snippet) > 400 {
						snippet = snippet[:400]
					}
					fmt.Println(string(snippet))
					// try to find an /N x 0 R inside the object (appearance ref)
					reN := regexp.MustCompile(`/N\s+(\d+)\s+0\s+R`)
					if nm := reN.FindSubmatch(objBytes); nm != nil {
						apNum := string(nm[1])
						// locate that object and try to extract stream
						hdr := []byte(fmt.Sprintf("%s 0 obj", apNum))
						if hdrPos := bytes.Index(b, hdr); hdrPos >= 0 {
							// find 'stream' after hdrPos
							streamPos := bytes.Index(b[hdrPos:], []byte("stream"))
							endStreamPos := bytes.Index(b[hdrPos:], []byte("endstream"))
							if streamPos >= 0 && endStreamPos >= 0 && endStreamPos > streamPos {
								streamStart := hdrPos + streamPos + len("stream")
								streamEnd := hdrPos + endStreamPos
								raw := b[streamStart:streamEnd]
								// trim leading newline
								if len(raw) > 0 && (raw[0] == '\n' || raw[0] == '\r') {
									raw = raw[1:]
								}
								// try flate decode
								r, err := zlib.NewReader(bytes.NewReader(raw))
								if err == nil {
									dec, err := ioutil.ReadAll(r)
									if err == nil {
										r.Close()
										fmt.Printf("   decoded AP (%s) first 400 bytes:\n", apNum)
										dsnip := dec
										if len(dsnip) > 400 {
											dsnip = dsnip[:400]
										}
										fmt.Println(string(dsnip))
										// print any literal parentheses text found
										litRe := regexp.MustCompile(`\(([^\)]*)\)\s*Tj`)
										if lit := litRe.FindSubmatch(dec); lit != nil {
											fmt.Printf("   text literal found in AP: '%s'\n", string(lit[1]))
										} else {
											// try TJ arrays
											litRe2 := regexp.MustCompile(`\[(?:[^\]]*)\]\s*TJ`)
											if lit2 := litRe2.FindSubmatch(dec); lit2 != nil {
												fmt.Printf("   TJ array found in AP (snippet): %s\n", string(lit2[0]))
											}
										}
									} else {
										fmt.Printf("   zlib read all error for AP %s: %v\n", apNum, err)
										r.Close()
									}
								} else {
									fmt.Printf("   zlib newreader error for AP %s: %v\n", apNum, err)
								}

							}
						}
					}
				} else {
					fmt.Printf("   containing object header: %s found at %d but endobj not found nearby\n", objNum, objPos)
				}
			} else {
				fmt.Printf("   containing object: not found near dictStart\n")
			}
		}

		fmt.Printf("\n")
	}

	// Report high-level AcroForm NeedAppearances
	acRe := regexp.MustCompile(`/AcroForm\s+(\d+)\s0\sR`)
	if am := acRe.FindSubmatch(b); len(am) > 1 {
		objNum := string(am[1])
		hdr := []byte(fmt.Sprintf("%s 0 obj", objNum))
		if objPos := bytes.Index(b, hdr); objPos >= 0 {
			// find dict
			dictStartRel := bytes.Index(b[objPos:], []byte("<<"))
			if dictStartRel >= 0 {
				ds := objPos + dictStartRel
				depth := 0
				i := ds
				dictEnd := -1
				for i < len(b)-1 {
					if i+1 < len(b) && b[i] == '<' && b[i+1] == '<' {
						depth++
						i += 2
						continue
					}
					if i+1 < len(b) && b[i] == '>' && b[i+1] == '>' {
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
					dictBytes := b[ds : dictEnd+2]
					if bytes.Contains(dictBytes, []byte("/NeedAppearances")) {
						fmt.Println("AcroForm dictionary contains /NeedAppearances (likely true/false shown inline). Check its value in the bytes snippet below:")
						// print short snippet
						start := ds
						if start+200 > len(b) {
							start = ds
						}
						end := dictEnd + 2
						if end > start+400 {
							end = start + 400
						}
						fmt.Println(string(dictBytes[:]))
					} else {
						fmt.Println("AcroForm dictionary: /NeedAppearances not present")
					}
				}
			}
		}
	}
}
