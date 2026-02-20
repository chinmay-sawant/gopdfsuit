package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// PageDetail represents the dimensions of a single PDF page with its number
type PageDetail struct {
	PageNum int     `json:"pageNum"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

// PageInfo contains metadata about PDF pages
type PageInfo struct {
	TotalPages int          `json:"totalPages"`
	Pages      []PageDetail `json:"pages"`
}

// TextPosition represents the position of text on a page
type TextPosition struct {
	Text   string  `json:"text"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// RedactionRect represents a region to redact
type RedactionRect struct {
	PageNum int     `json:"pageNum"` // 1-based
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

// GetPageInfo extracts page count and dimensions from a PDF
func GetPageInfo(pdfBytes []byte) (PageInfo, error) {
	if len(pdfBytes) == 0 {
		return PageInfo{}, errors.New("empty pdf bytes")
	}

	if trailerHasEncrypt(pdfBytes) {
		return PageInfo{}, errors.New("encrypted PDFs are not supported")
	}

	rootRef, ok := findRootRef(pdfBytes)
	if !ok {
		return PageInfo{}, errors.New("could not find PDF root object")
	}

	objMap, err := buildObjectMap(pdfBytes)
	if err != nil {
		return PageInfo{}, err
	}

	rootBody, ok := objMap[rootRef]
	if !ok {
		return PageInfo{}, errors.New("root object not found in map")
	}

	pagesRefRe := regexp.MustCompile(`/Pages\s+(\d+)\s+(\d+)\s+R`)
	pm := pagesRefRe.FindSubmatch(rootBody)
	if pm == nil {
		return PageInfo{}, errors.New("no /Pages reference in Root")
	}
	pagesKey := string(pm[1]) + " " + string(pm[2])

	var pageDims []PageDetail
	if err := traversePages(pagesKey, objMap, &pageDims); err != nil {
		return PageInfo{}, fmt.Errorf("error traversing page tree: %w", err)
	}

	return PageInfo{
		TotalPages: len(pageDims),
		Pages:      pageDims,
	}, nil
}

// ExtractTextPositions extracts text with coordinates from a specific page
func ExtractTextPositions(pdfBytes []byte, pageNum int) ([]TextPosition, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}

	objMap, err := buildObjectMap(pdfBytes)
	if err != nil {
		return nil, err
	}

	pageRef, err := findPageObject(objMap, pdfBytes, pageNum)
	if err != nil {
		return nil, err
	}

	pageBody := objMap[pageRef]
	contentBytes, err := extractPageContent(pageBody, objMap)
	if err != nil {
		return nil, err
	}

	// Simple text extraction logic
	// This is a simplified parser and might not handle all PDF complexity (rotations, complex encodings)
	return parseTextOperators(contentBytes), nil
}

// FindTextOccurrences searches for text across all pages and returns redaction rectangles
func FindTextOccurrences(pdfBytes []byte, searchText string) ([]RedactionRect, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	if searchText == "" {
		return nil, nil
	}

	info, err := GetPageInfo(pdfBytes)
	if err != nil {
		return nil, err
	}

	var redactions []RedactionRect
	searchText = strings.ToLower(searchText)

	for i := 1; i <= info.TotalPages; i++ {
		positions, err := ExtractTextPositions(pdfBytes, i)
		if err != nil {
			// Log error but continue? Or fail? Best to continue for search flexibility
			continue
		}

		for _, pos := range positions {
			// Case-insensitive match
			if strings.Contains(strings.ToLower(pos.Text), searchText) {
				redactions = append(redactions, RedactionRect{
					PageNum: i,
					X:       pos.X,
					Y:       pos.Y,
					Width:   pos.Width,
					Height:  pos.Height,
				})
			}
		}
	}
	return redactions, nil
}

// ApplyRedactions applies visual redaction rectangles to the PDF
func ApplyRedactions(pdfBytes []byte, redactions []RedactionRect) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	if len(redactions) == 0 {
		return pdfBytes, nil
	}

	objMap, err := buildObjectMap(pdfBytes)
	if err != nil {
		return nil, err
	}

	// Group redactions by page
	redactionsByPage := make(map[int][]RedactionRect)
	for _, r := range redactions {
		redactionsByPage[r.PageNum] = append(redactionsByPage[r.PageNum], r)
	}

	// Find highest object number
	maxObj := 0
	for k := range objMap {
		var n int
		_, _ = fmt.Sscanf(k, "%d", &n)
		if n > maxObj {
			maxObj = n
		}
	}
	nextObj := maxObj + 1

	// For each page with redactions, append a new content stream
	for pageNum, rects := range redactionsByPage {
		pageRef, err := findPageObject(objMap, pdfBytes, pageNum)
		if err != nil {
			return nil, fmt.Errorf("failed to find page %d: %w", pageNum, err)
		}
		pageBody := objMap[pageRef]

		// Create redaction stream content
		var sb strings.Builder
		sb.WriteString("q 0 0 0 rg ") // Save state, set black color
		for _, r := range rects {
			// Construct rectangle path: x y w h re f (fill)
			sb.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re f ", r.X, r.Y, r.Width, r.Height))
		}
		sb.WriteString("Q ") // Restore state
		streamContent := sb.String()

		// Create new stream object
		streamObjKey := fmt.Sprintf("%d 0", nextObj)
		nextObj++

		streamObj := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(streamContent), streamContent)
		objMap[streamObjKey] = []byte(streamObj)

		// Append this new object to the page's /Contents
		newPageBody := appendStreamToPage(pageBody, streamObjKey)
		objMap[pageRef] = newPageBody
	}

	return rebuildPDF(objMap, pdfBytes)
}

// buildObjectMap parses the PDF into a map of "num gen" -> body
func buildObjectMap(pdfBytes []byte) (map[string][]byte, error) {
	objMap := make(map[string][]byte)
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	matches := objRe.FindAllSubmatch(pdfBytes, -1)
	for _, m := range matches {
		key := string(m[1]) + " " + string(m[2])
		objMap[key] = m[3]
	}
	// TODO: Handle linearization/incremental updates or object streams if needed.
	// For now, this assumes standard PDF structure.
	return objMap, nil
}

func traversePages(key string, objMap map[string][]byte, dims *[]PageDetail) error {
	body, ok := objMap[key]
	if !ok {
		return nil
	}

	if bytesIndex(body, []byte("/Type /Page")) >= 0 || bytesIndex(body, []byte("/Type/Page")) >= 0 {
		mediaBox := extractMediaBox(body, objMap)
		width := mediaBox[2] - mediaBox[0]
		height := mediaBox[3] - mediaBox[1]
		*dims = append(*dims, PageDetail{
			PageNum: len(*dims) + 1,
			Width:   width,
			Height:  height,
		})
		return nil
	}

	if bytesIndex(body, []byte("/Type /Pages")) >= 0 || bytesIndex(body, []byte("/Type/Pages")) >= 0 {
		kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
		km := kidsRe.FindSubmatch(body)
		if km != nil {
			refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
			refs := refRe.FindAllSubmatch(km[1], -1)
			for _, r := range refs {
				kidKey := string(r[1]) + " " + string(r[2])
				if err := traversePages(kidKey, objMap, dims); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return nil
}

func extractMediaBox(body []byte, objMap map[string][]byte) [4]float64 {
	defaultBox := [4]float64{0, 0, 595.28, 841.89}
	rectRe := regexp.MustCompile(`/MediaBox\s*\[\s*([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s*\]`)
	match := rectRe.FindSubmatch(body)

	if match == nil {
		refRe := regexp.MustCompile(`/MediaBox\s+(\d+)\s+(\d+)\s+R`)
		refMatch := refRe.FindSubmatch(body)
		if refMatch != nil {
			refKey := string(refMatch[1]) + " " + string(refMatch[2])
			if refBody, ok := objMap[refKey]; ok {
				arrayRe := regexp.MustCompile(`\[\s*([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s*\]`)
				match = arrayRe.FindSubmatch(refBody)
			}
		}
	}

	if match != nil {
		x1, _ := strconv.ParseFloat(string(match[1]), 64)
		y1, _ := strconv.ParseFloat(string(match[2]), 64)
		x2, _ := strconv.ParseFloat(string(match[3]), 64)
		y2, _ := strconv.ParseFloat(string(match[4]), 64)
		return [4]float64{x1, y1, x2, y2}
	}
	return defaultBox
}

func findPageObject(objMap map[string][]byte, pdfBytes []byte, targetPage int) (string, error) {
	rootRef, ok := findRootRef(pdfBytes)
	if !ok {
		return "", errors.New("missing Root")
	}

	rootBody := objMap[rootRef]
	pagesRefRe := regexp.MustCompile(`/Pages\s+(\d+)\s+(\d+)\s+R`)
	pm := pagesRefRe.FindSubmatch(rootBody)
	if pm == nil {
		return "", errors.New("missing /Pages in Root")
	}
	pagesKey := string(pm[1]) + " " + string(pm[2])

	var foundKey string
	var currentPage int

	var walk func(string) error
	walk = func(key string) error {
		if foundKey != "" {
			return nil
		}
		body, ok := objMap[key]
		if !ok {
			return nil
		}

		if bytesIndex(body, []byte("/Type /Page")) >= 0 || bytesIndex(body, []byte("/Type/Page")) >= 0 {
			currentPage++
			if currentPage == targetPage {
				foundKey = key
			}
			return nil
		}

		if bytesIndex(body, []byte("/Type /Pages")) >= 0 || bytesIndex(body, []byte("/Type/Pages")) >= 0 {
			kidsRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
			km := kidsRe.FindSubmatch(body)
			if km != nil {
				refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
				refs := refRe.FindAllSubmatch(km[1], -1)
				for _, r := range refs {
					kidKey := string(r[1]) + " " + string(r[2])
					if err := walk(kidKey); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	if err := walk(pagesKey); err != nil {
		return "", err
	}

	if foundKey == "" {
		return "", fmt.Errorf("page %d not found", targetPage)
	}
	return foundKey, nil
}

func extractPageContent(pageBody []byte, objMap map[string][]byte) ([]byte, error) {
	contentsRe := regexp.MustCompile(`/Contents\s+(?:(\d+)\s+(\d+)\s+R|\[(.*?)\])`)
	match := contentsRe.FindSubmatch(pageBody)
	if match == nil {
		return nil, nil // Empty content
	}

	var contentKeys []string
	if len(match[1]) > 0 {
		// Single reference
		contentKeys = append(contentKeys, string(match[1])+" "+string(match[2]))
	} else if len(match[3]) > 0 {
		// Array of references
		refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
		refs := refRe.FindAllSubmatch(match[3], -1)
		for _, r := range refs {
			contentKeys = append(contentKeys, string(r[1])+" "+string(r[2]))
		}
	}

	var fullContent bytes.Buffer
	for _, key := range contentKeys {
		streamBody, ok := objMap[key]
		if !ok {
			continue
		}
		// Decompress stream
		streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
		sm := streamRe.FindSubmatch(streamBody)
		if sm != nil {
			raw := sm[1]
			dec, err := tryFlateDecompress(raw) // tryFlateDecompress from xfdf.go (shared package)
			if err != nil {
				dec, err = tryZlibDecompress(raw)
				if err != nil {
					dec = raw // Fallback
				}
			}
			fullContent.Write(dec)
			fullContent.WriteByte('\n') // Spacing
		}
	}
	return fullContent.Bytes(), nil
}

func parseTextOperators(content []byte) []TextPosition {
	// Very basic parser.
	// Iterate through content, find BT ... ET blocks.
	// Inside block, track Tm, and text showing operators.
	var positions []TextPosition

	strContent := string(content)
	// tokens := strings.Fields(strContent) // Unused
	// Use slightly better tokenizer or regex?
	// Regex is safer for "(...)" strings.

	// Extract BT...ET blocks first
	btEtRe := regexp.MustCompile(`(?s)BT(.*?)ET`)
	blocks := btEtRe.FindAllStringSubmatch(strContent, -1)

	for _, block := range blocks {
		inner := block[1]

		// Regex to find text showing operators: (string) Tj or [(string) 10 (string)] TJ
		// and positioning: Tx Ty Td or a b c d e f Tm
		// Note: Text matrix Tm is [a b 0 0 e f]. (e, f) = (x, y).
		// Td is relative move.

		// Simplified scan: look for (text) Tj and try to find preceding Tm or Td
		// This is heuristic and won't match complex layout perfectly.

		// Find strings: \( ( [^)]* ) \) \s* Tj
		tjRe := regexp.MustCompile(`\(([^)]*)\)\s*Tj`)
		matches := tjRe.FindAllStringSubmatchIndex(inner, -1)

		currentX, currentY := 0.0, 0.0

		// Find last Tm or Td before this match
		for _, m := range matches {
			text := m[2] // index of text content
			startPos := m[0]

			// Look backwards for Tm or Td
			preceding := inner[:startPos]

			// Find last Tm: 6 numbers followed by Tm
			tmRe := regexp.MustCompile(`([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+Tm`)
			tmMatches := tmRe.FindAllStringSubmatch(preceding, -1)
			if len(tmMatches) > 0 {
				lastTm := tmMatches[len(tmMatches)-1]
				currentX, _ = strconv.ParseFloat(lastTm[5], 64)
				currentY, _ = strconv.ParseFloat(lastTm[6], 64)
			}

			textStr := inner[text:m[3]]

			// Estimate width (heuristic: 5 points per char if font info missing)
			width := float64(len(textStr)) * 5.0
			height := 10.0 // heuristic

			positions = append(positions, TextPosition{
				Text:   textStr,
				X:      currentX,
				Y:      currentY,
				Width:  width,
				Height: height,
			})
		}
	}

	return positions
}

func appendStreamToPage(pageBody []byte, streamKey string) []byte {
	// Find /Contents
	contentsRe := regexp.MustCompile(`/Contents\s+(?:(\d+\s+\d+\s+R)|\[(.*?)\])`)
	match := contentsRe.FindSubmatchIndex(pageBody)

	refStr := fmt.Sprintf("%s R", streamKey)

	if match == nil {
		// No contents, add it
		// Need to insert before end of dict >>
		dictEnd := bytes.LastIndex(pageBody, []byte(">>"))
		if dictEnd == -1 {
			return pageBody // Corrupt?
		}
		var buf bytes.Buffer
		buf.Write(pageBody[:dictEnd])
		buf.WriteString(fmt.Sprintf("/Contents [%s]", refStr))
		buf.Write(pageBody[dictEnd:])
		return buf.Bytes()
	}

	// Has contents
	start, end := match[0], match[1]
	// original := pageBody[start:end] // unused

	var replacement string
	if match[2] != -1 && match[3] != -1 {
		// Single ref: /Contents 1 0 R -> /Contents [1 0 R newRef]
		oldRef := string(pageBody[match[2]:match[3]])
		replacement = fmt.Sprintf("/Contents [%s %s]", oldRef, refStr)
	} else if match[4] != -1 && match[5] != -1 {
		// Array: /Contents [1 0 R] -> /Contents [1 0 R newRef]
		oldContent := string(pageBody[match[4]:match[5]])
		replacement = fmt.Sprintf("/Contents [%s %s]", oldContent, refStr)
	}

	if replacement != "" {
		var buf bytes.Buffer
		buf.Write(pageBody[:start])
		buf.WriteString(replacement)
		buf.Write(pageBody[end:])
		return buf.Bytes()
	}

	return pageBody
}

func rebuildPDF(objMap map[string][]byte, originalBytes []byte) ([]byte, error) {
	// Simple rebuild similar to xfdf.go's logic but stripped down
	// 1. Write Header
	// 2. Write Objects (sorted)
	// 3. Write Xref
	// 4. Write Trailer

	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")

	// Collect keys and sort IDs
	type objMeta struct {
		id  int
		gen int
		key string
	}
	var objs []objMeta
	maxID := 0

	for k := range objMap {
		var id, gen int
		_, _ = fmt.Sscanf(k, "%d %d", &id, &gen)
		objs = append(objs, objMeta{id, gen, k})
		if id > maxID {
			maxID = id
		}
	}
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].id < objs[j].id
	})

	offsets := make(map[int]int)

	for _, o := range objs {
		offsets[o.id] = out.Len()
		body := objMap[o.key]

		out.WriteString(fmt.Sprintf("%d %d obj\n", o.id, o.gen))
		out.Write(body)
		// Ensure line break
		if !bytes.HasSuffix(body, []byte("\n")) {
			out.WriteString("\n")
		}
		if !bytes.HasSuffix(body, []byte("endobj\n")) && !bytes.HasSuffix(body, []byte("endobj")) {
			out.WriteString("endobj\n")
		} else if bytes.HasSuffix(body, []byte("endobj")) {
			out.WriteString("\n")
		}
	}

	xrefStart := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n", maxID+1))
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i <= maxID; i++ {
		if off, ok := offsets[i]; ok {
			out.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
		} else {
			out.WriteString("0000000000 65535 f \n")
		}
	}

	// Find Root ID
	rootRef, _ := findRootRef(originalBytes)
	rootID := 1
	_, _ = fmt.Sscanf(rootRef, "%d", &rootID)

	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", maxID+1, rootID, xrefStart))

	return out.Bytes(), nil
}
