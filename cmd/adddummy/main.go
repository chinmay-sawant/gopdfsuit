package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// CLI tool: insert /V (dummy) into widget annotation dictionaries that lack /V
func main() {
	inPath := flag.String("in", "../../sampledata/patient2/patient2.pdf", "input PDF path")
	outPath := flag.String("out", "../../sampledata/patient2/output.pdf", "output PDF path")
	flag.Parse()

	b, err := os.ReadFile(*inPath)
	if err != nil {
		log.Fatalf("failed to read input: %v", err)
	}

	// attempt to load an XFDF alongside the PDF (same base name)
	xfdfPath := strings.TrimSuffix(*inPath, ".pdf") + ".xfdf"
	xfdfMap, err := parseXfdf(xfdfPath)
	if err != nil {
		fmt.Printf("debug: couldn't parse xfdf %s: %v\n", xfdfPath, err)
		xfdfMap = map[string]string{}
	}

	inds := findWidgetDictRanges(b)
	fmt.Printf("debug: found %d candidate widget dict ranges\n", len(inds))
	if len(inds) == 0 {
		fmt.Println("No widget dictionaries found")
		// ensure viewers will regenerate appearances when needed
		b2 := addAcroFormIfMissing(b)
		bWithNA := ensureNeedAppearances(b2)
		if err := os.WriteFile(*outPath, bWithNA, 0644); err != nil {
			log.Fatalf("failed to write output: %v", err)
		}
		fmt.Printf("wrote copy to %s\n", *outPath)
		return
	}

	sort.Slice(inds, func(i, j int) bool { return inds[i].start > inds[j].start })
	out := b
	// compute highest existing object number so we can allocate new object
	// numbers for appearance streams without colliding
	objRe := regexp.MustCompile(`(?m)^(\d+)\s+0\s+obj`)
	matches := objRe.FindAllSubmatch(out, -1)
	highest := 0
	for _, m := range matches {
		if n, err := strconv.Atoi(string(m[1])); err == nil && n > highest {
			highest = n
		}
	}
	nextObj := highest + 1
	// collect appearance objects to append at the end
	type apObj struct {
		num    int
		data   []byte
		stream []byte // raw stream content (between stream\n and endstream)
	}
	var apObjs []apObj
	modified := 0
	for _, r := range inds {
		fmt.Printf("debug: candidate range start=%d end=%d\n", r.start, r.end)
		seg := out[r.start:r.end]
		if bytes.Contains(seg, []byte("/V ")) || bytes.Contains(seg, []byte("/V(")) || bytes.Contains(seg, []byte("/V/")) {
			continue
		}
		// determine field name (T) inside this widget dict and look up xfdf map
		dictBytes := out[r.start:r.end]
		fieldName := extractFieldName(dictBytes)
		val := ""
		if fieldName != "" {
			if v, ok := xfdfMap[fieldName]; ok {
				val = v
			}
		}
		// escape value for PDF literal string
		esc := escapePDFString(val)
		// choose literal string or name token depending on value
		var insertion []byte
		// detect if this widget is a button (checkbox/radio)
		isButton := bytes.Contains(dictBytes, []byte("/FT/Btn")) || bytes.Contains(dictBytes, []byte("/FT /Btn"))
		// prepare AP object for visible appearance for non-button fields
		apNum := -1
		if val != "" && !isButton {
			apNum = nextObj
			nextObj++
			// parse Rect to compute BBox
			x1, y1, x2, y2 := parseRect(dictBytes)
			w := x2 - x1
			h := y2 - y1
			if w <= 0 {
				w = 200
			}
			if h <= 0 {
				h = 14
			}
			// build simple appearance stream using /Helv font
			// position text slightly inset (2,2)
			// use 12 or available size from DA; keep 12 by default
			content := fmt.Sprintf("BT /Helv 12 Tf 0 g 2 2 Td (%s) Tj ET", esc)
			// make stream bytes and Form XObject
			stream := []byte(content)
			// compressing streams is optional; keep uncompressed for simplicity
			bbox := fmt.Sprintf("[0 0 %s %s]", formatFloat(w), formatFloat(h))
			objBytes := []byte(fmt.Sprintf("\n%d 0 obj\n<< /Type /XObject /Subtype /Form /BBox %s /Resources << /Font << /Helv << /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >> >> >> /Length %d >>\nstream\n%s\nendstream\nendobj\n", apNum, bbox, len(stream), stream))
			apObjs = append(apObjs, apObj{num: apNum, data: objBytes, stream: stream})
		}

		if isButton {
			// For button fields (checkbox/radio) set /V and /AS to a name
			// Determine the 'on' appearance name from the AP if present
			onName := "On"
			if bytes.Contains(dictBytes, []byte("/Yes")) {
				onName = "Yes"
			}
			if bytes.Contains(dictBytes, []byte("/On")) {
				onName = "On"
			}
			// interpret a variety of truthy values as checked
			lv := strings.ToLower(strings.TrimSpace(val))
			checked := lv == "yes" || lv == "y" || lv == "true" || lv == "1" || lv == "on" || lv == "checked"
			chosen := "Off"
			if checked {
				chosen = onName
			}
			// set value and appearance state as name tokens
			insertion = []byte(" /V /" + chosen + " /AS/" + chosen)
		} else {
			if isPDFName(val) {
				// name object (no parentheses)
				if apNum > 0 {
					insertion = []byte(" /V /" + val + fmt.Sprintf(" /AP<</N %d 0 R>>", apNum))
				} else {
					insertion = []byte(" /V /" + val)
				}
			} else {
				// escape value for PDF literal string
				if apNum > 0 {
					insertion = []byte(" /V (" + esc + ")" + fmt.Sprintf(" /AP<</N %d 0 R>>", apNum))
				} else {
					insertion = []byte(" /V (" + esc + ")")
				}
			}
		}

		// ensure widget has a DA (default appearance) so viewers use expected font
		if !bytes.Contains(dictBytes, []byte("/DA(")) {
			// prepend DA insertion as well
			insertion = append([]byte(" /DA(/Helv 12 Tf 0 g)"), insertion...)
		}

		// ensure widget field is read-only (set /Ff 1) so it's not editable/clickable
		if !bytes.Contains(dictBytes, []byte("/Ff")) {
			insertion = append([]byte(" /Ff 1"), insertion...)
		}

		// insert before the closing >> (r.end points just after the >>)
		insertPos := r.end - 2
		if insertPos < r.start {
			// sanity: fallback to r.end if computed insert position invalid
			insertPos = r.end
		}
		out = append(out[:insertPos], append(insertion, out[insertPos:]...)...)
		modified++
	}

	// Before appending appearance objects, remove any existing Form XObject objects
	// whose stream content exactly matches the appearance stream we're adding. This
	// avoids leaving a static Form XObject that draws the same text beneath the
	// widget appearance (which would produce a duplicate visual when the widget
	// is rendered).
	if len(apObjs) > 0 {
		for _, a := range apObjs {
			out = removeFormObjectsMatchingStream(out, a.stream)
		}

		sx := bytes.LastIndex(out, []byte("startxref"))
		for _, a := range apObjs {
			if sx >= 0 {
				out = append(out[:sx], append(a.data, out[sx:]...)...)
				// update sx to point before the same startxref for next insertion
				sx = sx + len(a.data)
			} else {
				out = append(out, a.data...)
			}
		}
	}

	out2 := addAcroFormIfMissing(out)
	out3 := ensureNeedAppearances(out2)
	if err := os.WriteFile(*outPath, out3, 0644); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
	fmt.Printf("wrote modified PDF to %s (modified %d widgets)\n", *outPath, modified)
}

// addAcroFormIfMissing ensures the Catalog (object 1) contains an /AcroForm
// reference. If absent it appends a minimal AcroForm object with
// /NeedAppearances true and an empty /Fields array, and references it from the Catalog.
func addAcroFormIfMissing(pdf []byte) []byte {
	if bytes.Contains(pdf, []byte("/AcroForm")) {
		return pdf
	}
	// find highest object number
	objRe := regexp.MustCompile(`(?m)^(\d+)\s+0\s+obj`)
	matches := objRe.FindAllSubmatch(pdf, -1)
	highest := 0
	for _, m := range matches {
		if n, err := strconv.Atoi(string(m[1])); err == nil && n > highest {
			highest = n
		}
	}
	next := highest + 1

	// find Catalog object (1 0 obj)
	objHeader := []byte("1 0 obj")
	objPos := bytes.Index(pdf, objHeader)
	if objPos < 0 {
		return pdf
	}
	dictStartRel := bytes.Index(pdf[objPos:], []byte("<<"))
	if dictStartRel < 0 {
		return pdf
	}
	dictStart := objPos + dictStartRel
	// find matching >>
	depth := 0
	i := dictStart
	dictEnd := -1
	for i < len(pdf)-1 {
		if i+1 < len(pdf) && pdf[i] == '<' && pdf[i+1] == '<' {
			depth++
			i += 2
			continue
		}
		if i+1 < len(pdf) && pdf[i] == '>' && pdf[i+1] == '>' {
			depth--
			i += 2
			if depth == 0 {
				dictEnd = i
				break
			}
			continue
		}
		i++
	}
	if dictEnd < 0 {
		return pdf
	}
	// insert AcroForm reference into Catalog dict
	acRef := []byte(fmt.Sprintf(" /AcroForm %d 0 R", next))
	pdf = append(pdf[:dictEnd], append(acRef, pdf[dictEnd:]...)...)

	// prepare minimal AcroForm object with a default resource (DR) mapping
	// providing standard fonts so viewers can render appearances without needing
	// to generate appearance streams for each widget.
	acObj := []byte(fmt.Sprintf("\n%d 0 obj\n<< /NeedAppearances true /Fields [] /DR << /Font << /Helv << /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >> /TiBI << /Type /Font /Subtype /Type1 /BaseFont /Times-BoldItalic /Encoding /WinAnsiEncoding >> >> >> >>\nendobj\n", next))
	// insert before startxref if present, else append
	sx := bytes.LastIndex(pdf, []byte("startxref"))
	if sx >= 0 {
		pdf = append(pdf[:sx], append(acObj, pdf[sx:]...)...)
	} else {
		pdf = append(pdf, acObj...)
	}
	return pdf
}

// ensureNeedAppearances sets /NeedAppearances true in the AcroForm dictionary
// if an AcroForm object is present. It returns the possibly-modified bytes.
func ensureNeedAppearances(pdf []byte) []byte {
	acRe := regexp.MustCompile(`/AcroForm\s+(\d+)\s0\sR`)
	am := acRe.FindSubmatch(pdf)
	if len(am) <= 1 {
		// no AcroForm reference found; nothing to do
		return pdf
	}
	objNum := string(am[1])
	objHeader := []byte(fmt.Sprintf("%s 0 obj", objNum))
	objPos := bytes.Index(pdf, objHeader)
	if objPos < 0 {
		return pdf
	}
	dictStartRel := bytes.Index(pdf[objPos:], []byte("<<"))
	if dictStartRel < 0 {
		return pdf
	}
	dictStart := objPos + dictStartRel
	// find matching >> using depth count
	depth := 0
	i := dictStart
	dictEnd := -1
	for i < len(pdf)-1 {
		if i+1 < len(pdf) && pdf[i] == '<' && pdf[i+1] == '<' {
			depth++
			i += 2
			continue
		}
		if i+1 < len(pdf) && pdf[i] == '>' && pdf[i+1] == '>' {
			depth--
			i += 2
			if depth == 0 {
				dictEnd = i
				break
			}
			continue
		}
		i++
	}
	if dictEnd < 0 {
		return pdf
	}
	dictBytes := pdf[dictStart:dictEnd]
	// ensure /NeedAppearances true inside the AcroForm dictionary
	needRe := regexp.MustCompile(`\s*/NeedAppearances\s+(?:true|false)`)
	if needRe.Match(dictBytes) {
		// replace occurrence just inside the AcroForm dict slice and write back
		newDict := needRe.ReplaceAll(dictBytes, []byte(" /NeedAppearances true"))
		pdf = append(pdf[:dictStart], append(newDict, pdf[dictEnd:]...)...)
	} else {
		// insert ' /NeedAppearances true' before dictEnd
		newDict := append(dictBytes, []byte(" /NeedAppearances true")...)
		pdf = append(pdf[:dictStart], append(newDict, pdf[dictEnd:]...)...)
	}

	// After updating NeedAppearances, ensure the AcroForm has a default resource
	// font dictionary (/DR) to map resource names like /Helv or /TiBI to actual
	// base fonts. If no /DR exists, insert one before the closing '>>'.
	// Re-locate the AcroForm object header and dict bounds because the PDF
	// byte offsets may have shifted after the previous edit.
	objHeader = []byte(fmt.Sprintf("%s 0 obj", objNum))
	objPos = bytes.Index(pdf, objHeader)
	if objPos < 0 {
		return pdf
	}
	dictStartRel = bytes.Index(pdf[objPos:], []byte("<<"))
	if dictStartRel < 0 {
		return pdf
	}
	dictStart = objPos + dictStartRel
	// find matching >> using depth count
	depth = 0
	i = dictStart
	dictEnd = -1
	for i < len(pdf)-1 {
		if i+1 < len(pdf) && pdf[i] == '<' && pdf[i+1] == '<' {
			depth++
			i += 2
			continue
		}
		if i+1 < len(pdf) && pdf[i] == '>' && pdf[i+1] == '>' {
			depth--
			i += 2
			if depth == 0 {
				dictEnd = i
				break
			}
			continue
		}
		i++
	}
	if dictEnd < 0 {
		return pdf
	}
	dictBytes = pdf[dictStart:dictEnd]
	if !bytes.Contains(dictBytes, []byte("/DR")) {
		drInsertion := []byte(" /DR << /Font << /Helv << /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >> /TiBI << /Type /Font /Subtype /Type1 /BaseFont /Times-BoldItalic /Encoding /WinAnsiEncoding >> >> >>")
		pdf = append(pdf[:dictEnd], append(drInsertion, pdf[dictEnd:]...)...)
	}
	return pdf
}

// isPDFName returns true when s is safe to write as a PDF name object (no spaces or delimiter chars)
func isPDFName(s string) bool {
	if s == "" {
		return false
	}
	// Only treat values that look like simple identifiers as PDF names.
	// This avoids interpreting dates like "2025-08-20" or values with hyphens
	// as name objects. Allow letters, digits and underscore, and require the
	// first character be a letter or underscore (no leading hyphen or digit).
	matched, _ := regexp.MatchString(`^[A-Za-z_][A-Za-z0-9_]*$`, s)
	return matched
}

type rng struct{ start, end int }

func findWidgetDictRanges(pdf []byte) []rng {
	needle := []byte("/Subtype/Widget")
	var ranges []rng
	searchPos := 0
	for {
		idx := bytes.Index(pdf[searchPos:], needle)
		if idx < 0 {
			break
		}
		absIdx := searchPos + idx
		// debug: show where needle found and a small surrounding context
		startCtx := absIdx - 40
		if startCtx < 0 {
			startCtx = 0
		}
		endCtx := absIdx + 40
		if endCtx > len(pdf) {
			endCtx = len(pdf)
		}
		fmt.Printf("debug: needle at %d context=%q\n", absIdx, pdf[startCtx:endCtx])
		// context window bounds (we'll scan backwards for the '<<' within this window)
		ctxStart := absIdx - 4096
		if ctxStart < 0 {
			ctxStart = 0
		}
		// only ctxStart is needed for backward scan limit

		// collect all '<<' positions in the context window and try each candidate
		var candidates []int
		slice := pdf[ctxStart : absIdx+1]
		off := 0
		for {
			r := bytes.Index(slice[off:], []byte("<<"))
			if r < 0 {
				break
			}
			candidates = append(candidates, ctxStart+off+r)
			off += r + 2
			if off >= len(slice) {
				break
			}
		}
		fmt.Printf("debug: candidates count=%d (nearest shown first)\n", len(candidates))
		if len(candidates) > 0 {
			// show up to 6 nearest candidate positions for context
			for i := len(candidates) - 1; i >= 0 && i >= len(candidates)-6; i-- {
				pos := candidates[i]
				ctxS := pos - 40
				if ctxS < 0 {
					ctxS = 0
				}
				ctxE := pos + 40
				if ctxE > len(pdf) {
					ctxE = len(pdf)
				}
				fmt.Printf("debug: candidate << at %d context=%q\n", pos, pdf[ctxS:ctxE])
			}
		}
		if len(candidates) == 0 {
			searchPos = absIdx + 1
			continue
		}

		// try candidates from nearest to farthest (reverse order)
		chosen := -1
		chosenEnd := -1
		for ci := len(candidates) - 1; ci >= 0; ci-- {
			cand := candidates[ci]
			// find matching >> for the dictionary from cand
			depth := 0
			j := cand
			var dictEnd int
			dictEnd = -1
			for j < len(pdf)-1 {
				if j+1 < len(pdf) && pdf[j] == '<' && pdf[j+1] == '<' {
					depth++
					j += 2
					continue
				}
				if j+1 < len(pdf) && pdf[j] == '>' && pdf[j+1] == '>' {
					depth--
					j += 2
					if depth == 0 {
						dictEnd = j
						break
					}
					continue
				}
				j++
			}
			// log candidate outcome
			if dictEnd < 0 {
				fmt.Printf("debug: cand=%d -> no matching >> found\n", cand)
				continue
			}
			if dictEnd <= absIdx {
				fmt.Printf("debug: cand=%d -> dictEnd=%d is before needle %d (skip)\n", cand, dictEnd, absIdx)
				continue
			}
			dictBytes := pdf[cand:dictEnd]
			// prepare snippet for logging
			snippet := dictBytes
			if len(snippet) > 200 {
				snippet = snippet[:200]
			}
			hasTypeAnnot := bytes.Contains(dictBytes, []byte("/Type/Annot")) || bytes.Contains(dictBytes, []byte("/Type /Annot"))
			fmt.Printf("debug: cand=%d dictEnd=%d hasTypeAnnot=%v dictSnippet=%q\n", cand, dictEnd, hasTypeAnnot, snippet)
			if !hasTypeAnnot {
				continue
			}
			chosen = cand
			chosenEnd = dictEnd
		}
		if chosen < 0 {
			searchPos = absIdx + 1
			continue
		}
		dictStart := chosen
		dictEnd := chosenEnd
		ranges = append(ranges, rng{start: dictStart, end: dictEnd})
		searchPos = absIdx + 1
	}

	return ranges
}

// (helper removed)

// parseXfdf reads a .xfdf file and returns a map[fieldName]value
func parseXfdf(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(bytes.NewReader(b))
	m := map[string]string{}
	type XField struct {
		XMLName xml.Name `xml:"field"`
		Name    string   `xml:"name,attr"`
		Value   string   `xml:"value"`
	}
	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF { // io isn't imported now; replace with nil check
				break
			}
			return nil, err
		}
		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "field" {
			var xf XField
			if err := dec.DecodeElement(&xf, &se); err != nil {
				// ignore individual decode errors
				continue
			}
			if xf.Name != "" {
				m[xf.Name] = xf.Value
			}
		}
	}
	return m, nil
}

// extractFieldName finds /T(fieldname) inside a dictionary blob and returns the fieldname (unescaped)
func extractFieldName(dict []byte) string {
	idx := bytes.Index(dict, []byte("/T("))
	if idx < 0 {
		return ""
	}
	start := idx + 3
	var sb strings.Builder
	for i := start; i < len(dict); i++ {
		b := dict[i]
		if b == ')' {
			return sb.String()
		}
		if b == '\\' && i+1 < len(dict) {
			sb.WriteByte(dict[i+1])
			i++
			continue
		}
		sb.WriteByte(b)
	}
	return sb.String()
}

// escapePDFString escapes parentheses and backslashes for a PDF literal string
func escapePDFString(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return r.Replace(s)
}

// parseRect extracts numeric values from a /Rect[...] entry inside a widget dict.
// It returns x1,y1,x2,y2 as floats; on failure returns zeros.
func parseRect(dict []byte) (float64, float64, float64, float64) {
	re := regexp.MustCompile(`/Rect\s*\[([^\]]+)\]`)
	m := re.FindSubmatch(dict)
	if len(m) < 2 {
		return 0, 0, 0, 0
	}
	parts := strings.Fields(string(m[1]))
	if len(parts) < 4 {
		return 0, 0, 0, 0
	}
	f := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}
	return f(parts[0]), f(parts[1]), f(parts[2]), f(parts[3])
}

func formatFloat(f float64) string {
	// simple formatting without trailing zeros
	s := strconv.FormatFloat(f, 'f', -1, 64)
	return s
}

// removeFormObjectsMatchingStream scans the PDF bytes for Form XObject objects
// (object dictionaries containing '/Type /XObject' and '/Subtype /Form') and
// removes any object whose stream body equals the provided streamBytes. Returns
// a new byte slice with matching objects removed.
func removeFormObjectsMatchingStream(pdf []byte, streamBytes []byte) []byte {
	// look for occurrences of "obj" headers and inspect following dict/stream
	objRe := regexp.MustCompile(`(?m)^(\d+)\s+0\s+obj`) // reuse earlier re
	matches := objRe.FindAllIndex(pdf, -1)
	if len(matches) == 0 {
		return pdf
	}

	// We'll build a new buffer skipping matched objects
	out := make([]byte, 0, len(pdf))
	last := 0
	for _, mi := range matches {
		objStart := mi[0]
		// attempt to find 'endobj' after objStart
		endobjIdx := bytes.Index(pdf[objStart:], []byte("endobj"))
		if endobjIdx < 0 {
			continue
		}
		objEnd := objStart + endobjIdx + len("endobj")
		objBytes := pdf[objStart:objEnd]
		// quick check for XObject /Subtype /Form
		if !bytes.Contains(objBytes, []byte("/Type /XObject")) || !bytes.Contains(objBytes, []byte("/Subtype /Form")) {
			continue
		}
		// find stream body
		streamIdx := bytes.Index(objBytes, []byte("stream\n"))
		if streamIdx < 0 {
			streamIdx = bytes.Index(objBytes, []byte("stream\r\n"))
		}
		if streamIdx < 0 {
			continue
		}
		streamStart := objStart + streamIdx + len("stream\n")
		// find endstream after streamStart
		endstreamIdx := bytes.Index(pdf[streamStart:], []byte("endstream"))
		if endstreamIdx < 0 {
			continue
		}
		streamEnd := streamStart + endstreamIdx
		body := pdf[streamStart:streamEnd]
		// normalize by trimming whitespace
		if bytes.Equal(bytes.TrimSpace(body), bytes.TrimSpace(streamBytes)) {
			// copy content before this object
			out = append(out, pdf[last:objStart]...)
			last = objEnd
		}
	}
	// append remainder
	out = append(out, pdf[last:]...)
	if len(out) == 0 {
		return pdf
	}
	return out
}
