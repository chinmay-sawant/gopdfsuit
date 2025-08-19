package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Widget represents a PDF widget annotation
type Widget struct {
	Rect  []float64
	Value string
	DA    string
}

// PageInfo contains information extracted from a PDF page object
type PageInfo struct {
	Body      string
	Annots    []int
	Contents  int
	Resources string
}

// AcroFormInfo contains AcroForm default appearance and font information
type AcroFormInfo struct {
	DA      string
	HelvRef int
}

// readBytes reads all bytes from a file
func readBytes(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// writeBytes writes bytes to a file
func writeBytes(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

// findObjects finds all PDF objects and returns them as a map
func findObjects(pdfBytes []byte) map[int][]byte {
	objs := make(map[int][]byte)
	re := regexp.MustCompile(`(?s)(\d+)\s+0\s+obj(.*?)endobj`)
	matches := re.FindAllSubmatch(pdfBytes, -1)

	for _, match := range matches {
		num, err := strconv.Atoi(string(match[1]))
		if err != nil {
			continue
		}
		body := strings.TrimSpace(string(match[2]))
		objs[num] = []byte(body)
	}
	return objs
}

// getTrailer extracts the PDF trailer
func getTrailer(pdfBytes []byte) []byte {
	re := regexp.MustCompile(`(?s)trailer(.*?)startxref`)
	match := re.FindSubmatch(pdfBytes)
	if match == nil {
		return []byte{}
	}
	return []byte(strings.TrimSpace(string(match[1])))
}

// extractPageInfo extracts information from a page object
func extractPageInfo(objs map[int][]byte, pageObjNum int) PageInfo {
	body := string(objs[pageObjNum])

	// Get Annots list of object refs
	var annots []int
	annotsRe := regexp.MustCompile(`(?s)/Annots\s*\[(.*?)\]`)
	if match := annotsRe.FindStringSubmatch(body); match != nil {
		refsRe := regexp.MustCompile(`(\d+)\s+0\s+R`)
		refs := refsRe.FindAllStringSubmatch(match[1], -1)
		for _, ref := range refs {
			if num, err := strconv.Atoi(ref[1]); err == nil {
				annots = append(annots, num)
			}
		}
	}

	// Find Contents ref
	contentsRe := regexp.MustCompile(`/Contents\s+(\d+)\s+0\s+R`)
	var contentsRef int
	if match := contentsRe.FindStringSubmatch(body); match != nil {
		contentsRef, _ = strconv.Atoi(match[1])
	}

	// Find Resources
	resRe := regexp.MustCompile(`(?s)/Resources\s*(<<.*?>>)`)
	var resources string
	if match := resRe.FindStringSubmatch(body); match != nil {
		resources = match[1]
	}

	return PageInfo{
		Body:      body,
		Annots:    annots,
		Contents:  contentsRef,
		Resources: resources,
	}
}

// parseWidget parses a widget annotation object
func parseWidget(objBody []byte) *Widget {
	s := string(objBody)

	// Only process /Subtype /Widget
	if !strings.Contains(s, "/Subtype /Widget") {
		return nil
	}

	// Parse Rect
	rectRe := regexp.MustCompile(`/Rect\s*\[\s*([\d\.-]+)\s+([\d\.-]+)\s+([\d\.-]+)\s+([\d\.-]+)\s*\]`)
	var rect []float64
	if match := rectRe.FindStringSubmatch(s); match != nil {
		for i := 1; i <= 4; i++ {
			if val, err := strconv.ParseFloat(match[i], 64); err == nil {
				rect = append(rect, val)
			}
		}
	}

	// Parse Value /V (literal string)
	var value string
	valueRe := regexp.MustCompile(`/V\s*\(`)
	if match := valueRe.FindStringIndex(s); match != nil {
		start := match[1] // index after the opening '('
		value = extractPDFLiteralString(s, start)
	}

	// Parse Default Appearance /DA
	var da string
	daRe := regexp.MustCompile(`/DA\s*\((.*?)\)`)
	if match := daRe.FindStringSubmatch(s); match != nil {
		da = match[1]
	}

	return &Widget{
		Rect:  rect,
		Value: value,
		DA:    da,
	}
}

// extractPDFLiteralString extracts a PDF literal string starting at the given position
func extractPDFLiteralString(s string, start int) string {
	var buf strings.Builder
	i := start
	length := len(s)

	for i < length {
		ch := s[i]
		if ch == ')' {
			break
		}
		if ch == '\\' && i+1 < length {
			esc := s[i+1]
			buf.WriteByte('\\')
			buf.WriteByte(esc)
			i += 2
		} else {
			buf.WriteByte(ch)
			i++
		}
	}

	return decodePDFLiteral(buf.String())
}

// decodePDFLiteral decodes PDF literal string escapes
func decodePDFLiteral(raw string) string {
	var out strings.Builder
	i := 0
	length := len(raw)

	for i < length {
		c := raw[i]
		if c != '\\' {
			out.WriteByte(c)
			i++
		} else {
			i++
			if i >= length {
				break
			}
			e := raw[i]
			switch e {
			case 'n':
				out.WriteByte('\n')
			case 'r':
				out.WriteByte('\r')
			case 't':
				out.WriteByte('\t')
			case 'b':
				out.WriteByte('\b')
			case 'f':
				out.WriteByte('\f')
			case '\\', '(', ')':
				out.WriteByte(e)
			default:
				if e >= '0' && e <= '7' {
					// Parse octal
					octal := string(e)
					i++
					for j := 0; j < 2 && i < length && raw[i] >= '0' && raw[i] <= '7'; j++ {
						octal += string(raw[i])
						i++
					}
					i-- // will be incremented at end of loop
					if val, err := strconv.ParseInt(octal, 8, 32); err == nil {
						out.WriteByte(byte(val))
					} else {
						out.WriteString(octal)
					}
				} else {
					out.WriteByte(e)
				}
			}
			i++
		}
	}

	return out.String()
}

// acroformDAAndFont extracts AcroForm default appearance and font information
func acroformDAAndFont(objs map[int][]byte) AcroFormInfo {
	for _, body := range objs {
		s := string(body)
		if strings.Contains(s, "/AcroForm") || strings.Contains(s, "/NeedAppearances") {
			var da string
			daRe := regexp.MustCompile(`/DA\s*\((.*?)\)`)
			if match := daRe.FindStringSubmatch(s); match != nil {
				da = match[1]
			}

			var fontRef int
			drRe := regexp.MustCompile(`(?s)/DR\s*(<<.*?>>)`)
			if match := drRe.FindStringSubmatch(s); match != nil {
				helvRe := regexp.MustCompile(`/Helv\s+(\d+)\s+0\s+R`)
				if helvMatch := helvRe.FindStringSubmatch(match[1]); helvMatch != nil {
					fontRef, _ = strconv.Atoi(helvMatch[1])
				}
			}

			return AcroFormInfo{DA: da, HelvRef: fontRef}
		}
	}
	return AcroFormInfo{}
}

// escapePDFText escapes text for PDF literal strings
func escapePDFText(t string) string {
	t = strings.ReplaceAll(t, "\\", "\\\\")
	t = strings.ReplaceAll(t, "(", "\\(")
	t = strings.ReplaceAll(t, ")", "\\)")
	return t
}

// buildTextDrawOps builds text drawing operations for the fields
func buildTextDrawOps(fields []*Widget, acroDA AcroFormInfo) []byte {
	var ops []string

	for _, f := range fields {
		if f.Value == "" || len(f.Rect) != 4 {
			continue
		}

		x0, y0, _, y1 := f.Rect[0], f.Rect[1], f.Rect[2], f.Rect[3]
		height := y1 - y0

		// Derive font size from DA
		size := 12.0
		if f.DA != "" {
			sizeRe := regexp.MustCompile(`(\d+(?:\.\d+)?)\s+Tf`)
			if match := sizeRe.FindStringSubmatch(f.DA); match != nil {
				if val, err := strconv.ParseFloat(match[1], 64); err == nil {
					size = val
				}
			}
		} else if acroDA.DA != "" {
			sizeRe := regexp.MustCompile(`(\d+(?:\.\d+)?)\s+Tf`)
			if match := sizeRe.FindStringSubmatch(acroDA.DA); match != nil {
				if val, err := strconv.ParseFloat(match[1], 64); err == nil {
					size = val
				}
			}
		}

		// Basic vertical centering and small left padding
		tx := x0 + 2
		ty := y0 + max(0, (height-size)/2.0)
		text := escapePDFText(f.Value)

		cmd := fmt.Sprintf("BT /Helv %.0f Tf 0 0 0 rg %.3f %.3f Td (%s) Tj ET",
			size, tx, ty, text)
		ops = append(ops, cmd)
	}

	if len(ops) == 0 {
		return []byte{}
	}

	content := "q\n" + strings.Join(ops, "\n") + "\nQ\n"
	return []byte(content)
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// mergeFontResources merges font resources
func mergeFontResources(pageResourcesText, fontObjText string) string {
	fontObjInner := strings.TrimSpace(fontObjText)
	fontObjInner = strings.TrimPrefix(fontObjInner, "<<")
	fontObjInner = strings.TrimSuffix(fontObjInner, ">>")
	fontObjInner = strings.TrimSpace(fontObjInner)

	procRe := regexp.MustCompile(`(?s)/ProcSet\s*(\[.*?\])`)
	var proc string
	if match := procRe.FindStringSubmatch(pageResourcesText); match != nil {
		proc = match[1]
	}

	fontEntries := fontObjInner
	newFont := "<< /Helv 5 0 R"
	if fontEntries != "" {
		newFont += " " + fontEntries
	}
	newFont += " >>"

	if proc != "" {
		return fmt.Sprintf("<< /Font %s /ProcSet %s >>", newFont, proc)
	}
	return fmt.Sprintf("<< /Font %s >>", newFont)
}

// rebuildPDF rebuilds the PDF with modifications
func rebuildPDF(objs map[int][]byte, modifiedObjs map[int][]byte, trailerRaw []byte, outPath string) error {
	// Build header
	header := []byte("%PDF-1.4\n%\xFF\xFF\xFF\xFF\n")

	// Find max object number
	maxObj := 0
	for num := range objs {
		if num > maxObj {
			maxObj = num
		}
	}
	for num := range modifiedObjs {
		if num > maxObj {
			maxObj = num
		}
	}

	var out []byte
	out = append(out, header...)
	offsets := make(map[int]int)
	offsets[0] = 0

	// Write objects
	for i := 1; i <= maxObj; i++ {
		offsets[i] = len(out)

		var body []byte
		if modBody, exists := modifiedObjs[i]; exists {
			body = modBody
		} else if origBody, exists := objs[i]; exists {
			body = origBody
		} else {
			body = []byte("<< /Type /Null >>")
		}

		objHeader := fmt.Sprintf("%d 0 obj\n", i)
		out = append(out, []byte(objHeader)...)
		out = append(out, body...)
		out = append(out, []byte("\nendobj\n")...)
	}

	// Write xref table
	xrefOffset := len(out)
	out = append(out, []byte("xref\n")...)
	out = append(out, []byte(fmt.Sprintf("0 %d\n", maxObj+1))...)
	out = append(out, []byte("0000000000 65535 f \n")...)

	for i := 1; i <= maxObj; i++ {
		out = append(out, []byte(fmt.Sprintf("%010d 00000 n \n", offsets[i]))...)
	}

	// Write trailer
	tr := string(trailerRaw)
	sizeRe := regexp.MustCompile(`/Size\s+\d+`)
	if sizeRe.MatchString(tr) {
		tr = sizeRe.ReplaceAllString(tr, fmt.Sprintf("/Size %d", maxObj+1))
	} else {
		tr = tr + fmt.Sprintf("\n/Size %d", maxObj+1)
	}

	out = append(out, []byte("trailer\n")...)
	out = append(out, []byte(tr)...)
	out = append(out, []byte("\nstartxref\n")...)
	out = append(out, []byte(fmt.Sprintf("%d\n", xrefOffset))...)
	out = append(out, []byte("%%EOF\n")...)

	return writeBytes(outPath, out)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run flatten_pdf.go input.pdf output.pdf")
		os.Exit(1)
	}

	inp := os.Args[1]
	outp := os.Args[2]

	pdf, err := readBytes(inp)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	objs := findObjects(pdf)
	trailer := getTrailer(pdf)

	// Find page object
	var pageNum int
	found := false
	for num, body := range objs {
		if strings.Contains(string(body), "/Type /Page") {
			pageNum = num
			found = true
			break
		}
	}

	if !found {
		fmt.Println("No page object found.")
		os.Exit(1)
	}

	pinfo := extractPageInfo(objs, pageNum)

	var fields []*Widget
	var annotsToKeep []int

	for _, a := range pinfo.Annots {
		if objBody, exists := objs[a]; exists {
			w := parseWidget(objBody)
			if w == nil {
				annotsToKeep = append(annotsToKeep, a)
				continue
			}

			// Check if it's a button
			s := string(objBody)
			isButton := strings.Contains(s, "/FT /Btn")

			if !isButton {
				// Check parent for button type
				parentRe := regexp.MustCompile(`/Parent\s+(\d+)\s+0\s+R`)
				if match := parentRe.FindStringSubmatch(s); match != nil {
					if pnum, err := strconv.Atoi(match[1]); err == nil {
						if parentBody, exists := objs[pnum]; exists {
							if strings.Contains(string(parentBody), "/FT /Btn") {
								isButton = true
							}
						}
					}
				}
			}

			if isButton {
				annotsToKeep = append(annotsToKeep, a)
			} else if w.Value != "" {
				fields = append(fields, w)
			}
		}
	}

	if len(fields) == 0 {
		fmt.Println("No non-button widget fields with values found; only buttons/checkboxes will be preserved.")
	}

	acro := acroformDAAndFont(objs)
	contentBytes := buildTextDrawOps(fields, acro)

	if len(contentBytes) == 0 {
		fmt.Println("No overlay content to add; writing original file copy.")
		if err := writeBytes(outp, pdf); err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create new content stream object
	nextObj := maxObj(objs) + 1
	streamBody := fmt.Sprintf("<< /Length %d >>\nstream\n", len(contentBytes))
	streamBodyBytes := append([]byte(streamBody), contentBytes...)
	streamBodyBytes = append(streamBodyBytes, []byte("\nendstream")...)

	// Modify page object
	pageBody := string(objs[pageNum])

	// Update Annots
	if len(annotsToKeep) > 0 {
		var annotsStrs []string
		for _, n := range annotsToKeep {
			annotsStrs = append(annotsStrs, fmt.Sprintf("%d 0 R", n))
		}
		annotsStr := strings.Join(annotsStrs, " ")
		annotsRe := regexp.MustCompile(`(?s)/Annots\s*\[.*?\]`)
		pageBody = annotsRe.ReplaceAllString(pageBody, fmt.Sprintf("/Annots [ %s ]", annotsStr))
	} else {
		annotsRe := regexp.MustCompile(`(?s)/Annots\s*\[.*?\]\s*`)
		pageBody = annotsRe.ReplaceAllString(pageBody, "")
	}

	// Update Contents
	contentsRe := regexp.MustCompile(`/Contents\s+\d+\s+0\s+R`)
	pageBody = contentsRe.ReplaceAllString(pageBody,
		fmt.Sprintf("/Contents [ %d 0 R %d 0 R ]", pinfo.Contents, nextObj))

	// Handle font resources if present
	if pinfo.Resources != "" {
		fontRe := regexp.MustCompile(`/Font\s+(\d+)\s+0\s+R`)
		if match := fontRe.FindStringSubmatch(pinfo.Resources); match != nil {
			if fontRefNum, err := strconv.Atoi(match[1]); err == nil {
				if fontObjBody, exists := objs[fontRefNum]; exists {
					newRes := mergeFontResources(pinfo.Resources, string(fontObjBody))
					resRe := regexp.MustCompile(`(?s)/Resources\s*<<.*?>>`)
					pageBody = resRe.ReplaceAllString(pageBody, "/Resources "+newRes)
				}
			}
		}
	}

	// Prepare modified objects
	modified := make(map[int][]byte)
	modified[nextObj] = streamBodyBytes
	modified[pageNum] = []byte(pageBody)

	// Rebuild PDF
	if err := rebuildPDF(objs, modified, trailer, outp); err != nil {
		fmt.Printf("Error rebuilding PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote flattened PDF to %s\n", outp)
}

// maxObj returns the maximum object number
func maxObj(objs map[int][]byte) int {
	max := 0
	for num := range objs {
		if num > max {
			max = num
		}
	}
	return max
}
