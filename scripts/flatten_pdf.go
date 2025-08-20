package scripts

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// This Go program is a direct conversion of the Python script `flatten_pdf.py`.
// It performs the same functions:
// - Parses objects from a single-file PDF.
// - Collects widget annotations from the first page.
// - For each widget with a value (/V), creates a page content stream to draw the text.
// - Removes the interactive form fields (/Annots) from the page.
// - Writes a new PDF with the added content stream and a rebuilt xref/trailer.
//
// Limitations from the original script are preserved:
// - Works on simple PDFs.
// - Does not decode or edit existing compressed content streams.
// - Best-effort flattening.
//
// Usage:
//   go run flatten_pdf.go input.pdf output.pdf

// Global compiled regex patterns for efficiency, mirroring those in the Python script.
var (
	objRegex              = regexp.MustCompile(`(?s)(\d+)\s+0\s+obj(.*?)endobj`)
	trailerRegex          = regexp.MustCompile(`(?s)trailer(.*?)startxref`)
	annotsRegex           = regexp.MustCompile(`(?s)/Annots\s*\[(.*?)\]`)
	refRegex              = regexp.MustCompile(`(\d+)\s+0\s+R`)
	contentsRegex         = regexp.MustCompile(`/Contents\s+(\d+)\s+0\s+R`)
	contentsArrRegex      = regexp.MustCompile(`/Contents\s*\[(.*?)\]`)
	resourcesRegex        = regexp.MustCompile(`(?s)/Resources\s*(<<.*?)>>`)
	rectRegex             = regexp.MustCompile(`/Rect\s*\[\s*([\d\.-]+)\s+([\d\.-]+)\s+([\d\.-]+)\s+([\d\.-]+)\s*\]`)
	valueStartRegex       = regexp.MustCompile(`/V\s*\(`)
	daRegex               = regexp.MustCompile(`/DA\s*\((.*?)\)`)
	valueNameRegex        = regexp.MustCompile(`/V\s*/([^\s\)<>\[]+)`)
	valueHexRegex         = regexp.MustCompile(`/V\s*<([0-9A-Fa-f]+)>`)
	valueRefRegex         = regexp.MustCompile(`/V\s+(\d+)\s+0\s+R`)
	parentRegex           = regexp.MustCompile(`/Parent\s+(\d+)\s+0\s+R`)
	drRegex               = regexp.MustCompile(`(?s)/DR\s*(<<.*?)>>`)
	helvRegex             = regexp.MustCompile(`/Helv\s+(\d+)\s+0\s+R`)
	fontSizeRegex         = regexp.MustCompile(`(\d+(?:\.\d+)?)\s+Tf`)
	fontRefRegex          = regexp.MustCompile(`/Font\s+(\d+)\s+0\s+R`)
	procSetRegex          = regexp.MustCompile(`(?s)/ProcSet\s*(\[.*?\])`)
	sizeRegex             = regexp.MustCompile(`/Size\s+\d+`)
	annotsReplaceRegex    = regexp.MustCompile(`(?s)/Annots\s*\[.*?\]\s*`)
	contentsReplaceRegex  = regexp.MustCompile(`/Contents\s+\d+\s+0\s+R`)
	resourcesReplaceRegex = regexp.MustCompile(`(?s)/Resources\s*<<.*?>>`)
)

// findObjects parses a PDF file's bytes and extracts all object bodies.
// It's equivalent to the Python `find_objects` function.
func findObjects(pdfBytes []byte) map[int][]byte {
	objs := make(map[int][]byte)
	matches := objRegex.FindAllSubmatch(pdfBytes, -1)
	// fmt.Printf("Found %d objects in PDF\n", len(matches))
	for _, m := range matches {
		num, err := strconv.Atoi(string(m[1]))
		if err != nil {
			// fmt.Printf("Warning: Failed to parse object number: %s\n", string(m[1]))
			continue
		}
		// In Go, trim whitespace from the byte slice.
		body := bytes.TrimSpace(m[2])
		objs[num] = body
		// fmt.Printf("Object %d: %d bytes\n", num, len(body))
	}
	return objs
}

// getTrailer extracts the trailer dictionary from the PDF bytes.
// It's equivalent to the Python `get_trailer` function.
func getTrailer(pdfBytes []byte) []byte {
	m := trailerRegex.FindSubmatch(pdfBytes)
	if len(m) > 1 {
		return bytes.TrimSpace(m[1])
	}
	return nil
}

// PageInfo holds extracted information from a Page object.
type PageInfo struct {
	Body      string
	Annots    []int
	Contents  int
	Resources string
}

// extractPageInfo parses a page object to find annotations, contents, and resources.
// It's equivalent to the Python `extract_page_info` function.
func extractPageInfo(objs map[int][]byte, pageObjNum int) (*PageInfo, error) {
	bodyBytes, ok := objs[pageObjNum]
	if !ok {
		return nil, fmt.Errorf("page object %d not found", pageObjNum)
	}
	// The Python script uses 'latin1' encoding, which is a 1-to-1 byte mapping.
	// In Go, we can convert the byte slice to a string, as it will preserve the bytes.
	body := string(bodyBytes)

	info := &PageInfo{Body: body}

	// Get Annots list of object refs
	annotsMatch := annotsRegex.FindStringSubmatch(body)
	if len(annotsMatch) > 1 {
		refsMatches := refRegex.FindAllStringSubmatch(annotsMatch[1], -1)
		for _, r := range refsMatches {
			if num, err := strconv.Atoi(r[1]); err == nil {
				info.Annots = append(info.Annots, num)
			}
		}
	}

	// Find Contents ref (single) or array; prefer the first ref found
	contentsMatch := contentsRegex.FindStringSubmatch(body)
	if len(contentsMatch) > 1 {
		if num, err := strconv.Atoi(contentsMatch[1]); err == nil {
			info.Contents = num
		}
	} else {
		contentsArrMatch := contentsArrRegex.FindStringSubmatch(body)
		if len(contentsArrMatch) > 1 {
			refs := refRegex.FindAllStringSubmatch(contentsArrMatch[1], -1)
			if len(refs) > 0 {
				if num, err := strconv.Atoi(refs[0][1]); err == nil {
					info.Contents = num
				}
			}
		}
	}

	// Find Resources
	resourcesMatch := resourcesRegex.FindStringSubmatch(body)
	if len(resourcesMatch) > 1 {
		info.Resources = resourcesMatch[1]
	}

	return info, nil
}

// WidgetInfo holds extracted information from a Widget annotation object.
type WidgetInfo struct {
	Rect  []float64
	Value string
	DA    string
}

// decodePDFLiteral decodes a PDF literal string, handling escapes like \(, \), \\, and octal codes.
func decodePDFLiteral(s string) string {
	var out strings.Builder
	r := strings.NewReader(s)
	for r.Len() > 0 {
		ch, _, _ := r.ReadRune()
		if ch != '\\' {
			out.WriteRune(ch)
			continue
		}
		// Handle escape sequence
		if r.Len() == 0 {
			break
		}
		esc, _, _ := r.ReadRune()
		switch esc {
		case 'n':
			out.WriteRune('\n')
		case 'r':
			out.WriteRune('\r')
		case 't':
			out.WriteRune('\t')
		case 'b':
			out.WriteRune('\b')
		case 'f':
			out.WriteRune('\f')
		case '\\', '(', ')':
			out.WriteRune(esc)
		default:
			if esc >= '0' && esc <= '7' {
				// Octal escape
				oct := string(esc)
				for i := 0; i < 2 && r.Len() > 0; i++ {
					peek, _, _ := r.ReadRune()
					if peek >= '0' && peek <= '7' {
						oct += string(peek)
					} else {
						r.UnreadRune()
						break
					}
				}
				if val, err := strconv.ParseInt(oct, 8, 32); err == nil {
					out.WriteRune(rune(val))
				} else {
					out.WriteString(oct) // fallback
				}
			} else {
				out.WriteRune(esc) // Unknown escape
			}
		}
	}
	return out.String()
}

// parseWidget parses a widget annotation's body.
// It's equivalent to the Python `parse_widget` function.
func parseWidget(objBody []byte, objs map[int][]byte) *WidgetInfo {
	s := string(objBody)
	if !strings.Contains(s, "/Subtype /Widget") {
		return nil
	}

	// fmt.Printf("Parsing widget: %s\n", s[:min(200, len(s))])

	w := &WidgetInfo{}

	// Rect
	rectMatch := rectRegex.FindStringSubmatch(s)
	if len(rectMatch) == 5 {
		for i := 1; i <= 4; i++ {
			if f, err := strconv.ParseFloat(rectMatch[i], 64); err == nil {
				w.Rect = append(w.Rect, f)
			}
		}
		// fmt.Printf("Found rect: %v\n", w.Rect)
	}

	// Value /V could be a literal, a name (/Name), a hex string (<...>), or an indirect ref.
	// Try literal first
	if loc := valueStartRegex.FindStringIndex(s); loc != nil {
		// loc[1] points after the '(' so pass index of '('
		start := loc[1] - 1
		val, _ := extractLiteralFrom(s, start)
		w.Value = val
		// fmt.Printf("Found literal value: '%s'\n", val)
	} else if m := valueNameRegex.FindStringSubmatch(s); len(m) > 1 {
		w.Value = m[1]
		// fmt.Printf("Found name value: '%s'\n", w.Value)
	} else if m := valueHexRegex.FindStringSubmatch(s); len(m) > 1 {
		if bs, err := hexDecode(m[1]); err == nil {
			w.Value = string(bs)
			// fmt.Printf("Found hex value: '%s'\n", w.Value)
		}
	} else if m := valueRefRegex.FindStringSubmatch(s); len(m) > 1 {
		if refNum, err := strconv.Atoi(m[1]); err == nil {
			w.Value = resolveValueFromObj(objs, refNum)
			// fmt.Printf("Found ref value: '%s' (from obj %d)\n", w.Value, refNum)
		}
	}

	// Default Appearance /DA
	daMatch := daRegex.FindStringSubmatch(s)
	if len(daMatch) > 1 {
		w.DA = daMatch[1]
	}

	return w
}

// extractLiteralFrom finds a literal string starting at the given index of the input
// (where the '(' is located) and decodes it handling escapes. It returns the
// decoded string and the position after the closing ')'.
func extractLiteralFrom(s string, startIdx int) (string, int) {
	var buf strings.Builder
	i := startIdx + 1 // skip '('
	for i < len(s) {
		ch := s[i]
		if ch == ')' {
			return decodePDFLiteral(buf.String()), i + 1
		}
		if ch == '\\' {
			if i+1 < len(s) {
				// copy the escaped char sequence into buf so decodePDFLiteral can process
				buf.WriteByte('\\')
				buf.WriteByte(s[i+1])
				i += 2
				continue
			}
			buf.WriteByte('\\')
			i++
			continue
		}
		buf.WriteByte(ch)
		i++
	}
	// no closing paren
	return decodePDFLiteral(buf.String()), i
}

// resolveValueFromObj attempts to resolve a /V that is an indirect reference
// by reading the referenced object and extracting a literal, hex, or name value.
func resolveValueFromObj(objs map[int][]byte, ref int) string {
	body, ok := objs[ref]
	if !ok {
		return ""
	}
	s := string(body)
	// Try literal inside the object
	if loc := valueStartRegex.FindStringIndex(s); loc != nil {
		// start points to the '('
		start := loc[1] - 1
		val, _ := extractLiteralFrom(s, start)
		return val
	}
	// Try hex string
	if m := valueHexRegex.FindStringSubmatch(s); len(m) > 1 {
		if b, err := strconv.ParseUint("0", 10, 64); err == nil { // noop to keep imports
			_ = b
		}
		if bs, err := hexDecode(m[1]); err == nil {
			return string(bs)
		}
	}
	// Try name
	if m := valueNameRegex.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	// As a last resort, if the object body itself is a literal string (starts with '(')
	ts := strings.TrimSpace(s)
	if len(ts) > 0 && ts[0] == '(' {
		val, _ := extractLiteralFrom(ts, 0)
		return val
	}
	return ""
}

// hexDecode decodes a hex string (without angle brackets) into bytes.
func hexDecode(h string) ([]byte, error) {
	// If odd length, pad with a 0
	if len(h)%2 == 1 {
		h = "0" + h
	}
	dst := make([]byte, len(h)/2)
	for i := 0; i < len(dst); i++ {
		v, err := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, err
		}
		dst[i] = byte(v)
	}
	return dst, nil
}

// AcroFormInfo holds information from the AcroForm dictionary.
type AcroFormInfo struct {
	DA      string
	HelvRef int
}

// acroformDAAndFont finds the AcroForm dictionary and extracts default appearance and font info.
func acroformDAAndFont(objs map[int][]byte) *AcroFormInfo {
	info := &AcroFormInfo{}
	for _, body := range objs {
		s := string(body)
		if strings.Contains(s, "/AcroForm") || strings.Contains(s, "/NeedAppearances") {
			daMatch := daRegex.FindStringSubmatch(s)
			if len(daMatch) > 1 {
				info.DA = daMatch[1]
			}

			drMatch := drRegex.FindStringSubmatch(s)
			if len(drMatch) > 1 {
				helvMatch := helvRegex.FindStringSubmatch(drMatch[1])
				if len(helvMatch) > 1 {
					if ref, err := strconv.Atoi(helvMatch[1]); err == nil {
						info.HelvRef = ref
					}
				}
			}
			return info
		}
	}
	return info
}

// escapePDFText escapes characters for a PDF literal string.
func escapePDFText(t string) string {
	r := strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`)
	return r.Replace(t)
}

// buildTextDrawOps creates the content stream to draw flattened field values.
// Equivalent to the Python `build_text_draw_ops` function.
func buildTextDrawOps(fields []*WidgetInfo, acroDA *AcroFormInfo) []byte {
	var ops []string
	for _, f := range fields {
		if f.Value == "" || len(f.Rect) != 4 {
			continue
		}
		x0, y0, _, y1 := f.Rect[0], f.Rect[1], f.Rect[2], f.Rect[3]
		// width := x1 - x0 // This variable was unused. REMOVED.
		height := y1 - y0

		size := 12.0
		daString := f.DA
		if daString == "" && acroDA != nil {
			daString = acroDA.DA
		}
		if daString != "" {
			m := fontSizeRegex.FindStringSubmatch(daString)
			if len(m) > 1 {
				if s, err := strconv.ParseFloat(m[1], 64); err == nil {
					size = s
				}
			}
		}

		tx := x0 + 2
		ty := y0 + (height-size)/2.0
		if ty < y0 {
			ty = y0
		}
		text := escapePDFText(f.Value)
		cmd := fmt.Sprintf("BT /Helv %.2f Tf 0 0 0 rg %.3f %.3f Td (%s) Tj ET", size, tx, ty, text)
		ops = append(ops, cmd)
	}

	if len(ops) == 0 {
		return nil
	}

	content := "q\n" + strings.Join(ops, "\n") + "\nQ\n"
	return []byte(content)
}

// mergeFontResources merges existing font dictionaries with the one needed for flattening.
// Equivalent to the Python `merge_font_resources` function.
func mergeFontResources(pageResourcesText, fontObjText string) string {
	fontObjInner := strings.TrimSpace(fontObjText)
	fontObjInner = strings.TrimPrefix(fontObjInner, "<<")
	fontObjInner = strings.TrimSuffix(fontObjInner, ">>")
	fontObjInner = strings.TrimSpace(fontObjInner)

	procMatch := procSetRegex.FindStringSubmatch(pageResourcesText)
	var proc string
	if len(procMatch) > 1 {
		proc = procMatch[1]
	}

	newFontDict := "<< /Helv 5 0 R"
	if fontObjInner != "" {
		newFontDict += " " + fontObjInner
	}
	newFontDict += " >>"

	if proc != "" {
		return fmt.Sprintf("<< /Font %s /ProcSet %s >>", newFontDict, proc)
	}
	return fmt.Sprintf("<< /Font %s >>", newFontDict)
}

// rebuildPDF constructs the new PDF file from original and modified objects.
// Equivalent to the Python `rebuild_pdf` function.
func rebuildPDF(objs map[int][]byte, modifiedObjs map[int][]byte, trailerRaw []byte) ([]byte, error) {
	var out bytes.Buffer
	header := "%PDF-1.4\n%\xff\xff\xff\xff\n"
	out.WriteString(header)

	maxObj := 0
	for k := range objs {
		if k > maxObj {
			maxObj = k
		}
	}
	for k := range modifiedObjs {
		if k > maxObj {
			maxObj = k
		}
	}

	offsets := make([]int, maxObj+1)
	for i := 1; i <= maxObj; i++ {
		offsets[i] = out.Len()
		body, ok := modifiedObjs[i]
		if !ok {
			body, ok = objs[i]
		}
		if !ok {
			// Emit null object to preserve numbering
			body = []byte("<< /Type /Null >>")
		}

		fmt.Fprintf(&out, "%d 0 obj\n", i)
		out.Write(body)
		out.WriteString("\nendobj\n")
	}

	xrefOffset := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", maxObj+1)
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i <= maxObj; i++ {
		fmt.Fprintf(&out, "%010d 00000 n \n", offsets[i])
	}

	// Update trailer size
	tr := string(trailerRaw)
	newSize := fmt.Sprintf("/Size %d", maxObj+1)
	if strings.Contains(tr, "/Size") {
		tr = sizeRegex.ReplaceAllString(tr, newSize)
	} else {
		tr += "\n" + newSize
	}

	out.WriteString("trailer\n")
	out.WriteString(tr)
	fmt.Fprintf(&out, "\nstartxref\n%d\n%%%%EOF\n", xrefOffset)

	return out.Bytes(), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func FlattenPDFBytes(pdf []byte) ([]byte, error) {
	// fmt.Printf("Read %d bytes from PDF\n", len(pdf))

	objs := findObjects(pdf)
	trailer := getTrailer(pdf)

	// Find all page objects and their annotation counts
	pageNumbers := make([]int, 0)
	pageAnnotCounts := make(map[int]int)

	fmt.Println("Looking for page objects...")
	for num, body := range objs {
		if bytes.Contains(body, []byte("/Type /Page")) {
			pageNumbers = append(pageNumbers, num)

			// Count annotations for this page
			pinfo, err := extractPageInfo(objs, num)
			if err != nil {
				// fmt.Printf("Error extracting info for page %d: %v\n", num, err)
				continue
			}
			pageAnnotCounts[num] = len(pinfo.Annots)
			// fmt.Printf("Found page object: %d with %d annotations\n", num, len(pinfo.Annots))
		}
	}

	if len(pageNumbers) == 0 {
		return nil, fmt.Errorf("no page object found")
	}

	// Choose the page with the most annotations (most likely to be the form page)
	// If tied, choose the lowest numbered page for consistency
	pageNum := pageNumbers[0]
	maxAnnots := pageAnnotCounts[pageNum]

	for _, num := range pageNumbers {
		if pageAnnotCounts[num] > maxAnnots ||
			(pageAnnotCounts[num] == maxAnnots && num < pageNum) {
			pageNum = num
			maxAnnots = pageAnnotCounts[num]
		}
	}

	// fmt.Printf("Selected page object %d with %d annotations\n", pageNum, maxAnnots)

	pinfo, err := extractPageInfo(objs, pageNum)
	if err != nil {
		log.Fatalf("Error extracting page info: %v", err)
		return nil, err
	}

	// fmt.Printf("Page has %d annotations\n", len(pinfo.Annots))

	var fields []*WidgetInfo
	var annotsToKeep []int

	for _, a := range pinfo.Annots {
		raw, ok := objs[a]
		if !ok {
			// fmt.Printf("Warning: Annotation object %d not found\n", a)
			continue
		}

		// fmt.Printf("Processing annotation %d\n", a)
		w := parseWidget(raw, objs)
		if w == nil {
			// fmt.Printf("Annotation %d is not a widget, keeping\n", a)
			annotsToKeep = append(annotsToKeep, a)
			continue
		}

		// Check if the widget is a button/checkbox to preserve it
		isButton := false
		s := string(raw)
		if strings.Contains(s, "/FT /Btn") {
			isButton = true
			// fmt.Printf("Widget %d is a button (direct)\n", a)
		} else {
			pm := parentRegex.FindStringSubmatch(s)
			if len(pm) > 1 {
				pnum, _ := strconv.Atoi(pm[1])
				if pbody, pexists := objs[pnum]; pexists && bytes.Contains(pbody, []byte("/FT /Btn")) {
					isButton = true
					// fmt.Printf("Widget %d is a button (inherited from parent %d)\n", a, pnum)
				}
			}
		}

		if isButton {
			// fmt.Printf("Keeping button widget %d\n", a)
			annotsToKeep = append(annotsToKeep, a)
		} else if w.Value != "" {
			// fmt.Printf("Adding text field %d with value: '%s'\n", a, w.Value)
			fields = append(fields, w)
		} else {
			// fmt.Printf("Widget %d has no value, skipping\n", a)
		}
	}

	// fmt.Printf("Found %d text fields with values, %d annotations to keep\n", len(fields), len(annotsToKeep))

	if len(fields) == 0 {
		fmt.Println("No non-button widget fields with values found; only buttons/checkboxes will be preserved.")
	}

	acro := acroformDAAndFont(objs)
	contentBytes := buildTextDrawOps(fields, acro)

	if len(contentBytes) == 0 {
		fmt.Println("No overlay content to add; writing original file copy.")
		return pdf, nil
	}

	// fmt.Printf("Generated %d bytes of content stream\n", len(contentBytes))

	// Create new content stream object
	nextObj := 0
	for k := range objs {
		if k > nextObj {
			nextObj = k
		}
	}
	nextObj++

	streamBody := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(contentBytes), contentBytes)

	// Modify page object
	pageBody := pinfo.Body
	if len(annotsToKeep) > 0 {
		var annotsRefs []string
		for _, n := range annotsToKeep {
			annotsRefs = append(annotsRefs, fmt.Sprintf("%d 0 R", n))
		}
		newAnnots := fmt.Sprintf("/Annots [ %s ]", strings.Join(annotsRefs, " "))
		pageBody = annotsRegex.ReplaceAllString(pageBody, newAnnots)
	} else {
		pageBody = annotsReplaceRegex.ReplaceAllString(pageBody, "")
	}
	newContents := fmt.Sprintf("/Contents [ %d 0 R %d 0 R ]", pinfo.Contents, nextObj)
	pageBody = contentsReplaceRegex.ReplaceAllString(pageBody, newContents)

	// Merge font resources
	if pinfo.Resources != "" {
		m := fontRefRegex.FindStringSubmatch(pinfo.Resources)
		if len(m) > 1 {
			fontRefNum, _ := strconv.Atoi(m[1])
			if fontObjBytes, ok := objs[fontRefNum]; ok {
				newRes := mergeFontResources(pinfo.Resources, string(fontObjBytes))
				pageBody = resourcesReplaceRegex.ReplaceAllString(pageBody, "/Resources "+newRes)
			}
		}
	}

	modified := make(map[int][]byte)
	modified[nextObj] = []byte(streamBody)
	modified[pageNum] = []byte(pageBody)

	// Rebuild and write PDF
	var outBytes []byte
	outBytes, err = rebuildPDF(objs, modified, trailer)
	if err != nil {
		log.Fatalf("Failed to rebuild PDF: %v", err)
		return nil, err
	}
	return outBytes, nil
}
