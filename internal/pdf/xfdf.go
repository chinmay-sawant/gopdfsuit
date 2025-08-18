package pdf

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

// DetectFormFields scans PDF bytes and returns a unique list of AcroForm field names.
// This is a heuristic: it looks for occurrences of "/T (name)" inside the PDF bytes.
func DetectFormFields(pdfBytes []byte) ([]string, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	// Regex to find /T (fieldname)
	re := regexp.MustCompile(`/T\s*\(([^\)]*)\)`)
	matches := re.FindAllSubmatch(pdfBytes, -1)
	set := make(map[string]struct{})
	for _, m := range matches {
		if len(m) > 1 {
			name := string(m[1])
			if strings.TrimSpace(name) != "" {
				set[name] = struct{}{}
			}
		}
	}
	var names []string
	for k := range set {
		names = append(names, k)
	}
	return names, nil
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
		return out, nil
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

	return out, nil
}
