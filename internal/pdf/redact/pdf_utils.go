package redact

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// buildObjectMap parses the PDF into a map of "num gen" -> body
func buildObjectMap(pdfBytes []byte) (map[string][]byte, error) {
	objMap := make(map[string][]byte)
	objRe := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj(.*?)endobj`)
	matches := objRe.FindAllSubmatch(pdfBytes, -1)
	for _, m := range matches {
		key := string(m[1]) + " " + string(m[2])
		body := m[3]

		// Expand object streams so downstream page/content lookups work on modern PDFs.
		if bytesIndex(body, []byte("/ObjStm")) >= 0 || bytesIndex(body, []byte("/Type/ObjStm")) >= 0 {
			streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
			if sm := streamRe.FindSubmatch(body); sm != nil {
				streamBytes := sm[1]
				var dec []byte
				if d, err := tryZlibDecompress(streamBytes); err == nil {
					dec = d
				} else if d, err := tryFlateDecompress(streamBytes); err == nil {
					dec = d
				}
				if dec != nil {
					firstRe := regexp.MustCompile(`/First\s+(\d+)`)
					first := 0
					if fm := firstRe.FindSubmatch(body); fm != nil {
						if _, err := fmt.Sscanf(string(fm[1]), "%d", &first); err != nil {
							first = 0
						}
					}
					if first > 0 && first < len(dec) {
						header := strings.TrimSpace(string(dec[:first]))
						parts := strings.Fields(header)
						content := dec[first:]
						for i := 0; i+1 < len(parts); i += 2 {
							var objNum, off int
							if _, err := fmt.Sscanf(parts[i], "%d", &objNum); err != nil {
								continue
							}
							if _, err := fmt.Sscanf(parts[i+1], "%d", &off); err != nil {
								continue
							}
							end := len(content)
							for j := i + 2; j+1 < len(parts); j += 2 {
								var nextOff int
								if _, err := fmt.Sscanf(parts[j+1], "%d", &nextOff); err == nil {
									end = nextOff
									break
								}
							}
							if off < 0 || off >= len(content) || end <= off || end > len(content) {
								continue
							}
							objMap[fmt.Sprintf("%d 0", objNum)] = content[off:end]
						}
						objMap[key] = body
						continue
					}
				}
			}
		}

		objMap[key] = body
	}
	parseXRefStreams(pdfBytes, objMap)
	// TODO: Handle linearization/incremental updates or object streams if needed.
	// For now, this assumes standard PDF structure.
	return objMap, nil
}

func traversePages(key string, objMap map[string][]byte, dims *[]models.PageDetail) error {
	body, ok := objMap[key]
	if !ok {
		return nil
	}

	if isPDFTypePages(body) {
		for _, kidKey := range extractKidsRefs(body) {
			if err := traversePages(kidKey, objMap, dims); err != nil {
				return err
			}
		}
		return nil
	}

	if isPDFTypePage(body) {
		mediaBox := extractMediaBox(body, objMap)
		width := mediaBox[2] - mediaBox[0]
		height := mediaBox[3] - mediaBox[1]
		*dims = append(*dims, models.PageDetail{
			PageNum: len(*dims) + 1,
			Width:   width,
			Height:  height,
		})
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

		if isPDFTypePages(body) {
			for _, kidKey := range extractKidsRefs(body) {
				if err := walk(kidKey); err != nil {
					return err
				}
			}
			return nil
		}

		if isPDFTypePage(body) {
			currentPage++
			if currentPage == targetPage {
				foundKey = key
			}
			return nil
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
	resources := findPageResources(pageBody, objMap)

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
	visitedXObjects := map[string]bool{}
	for _, key := range contentKeys {
		streamBody, ok := objMap[key]
		if !ok {
			continue
		}
		start, end, ok := locateStreamSegment(streamBody)
		if ok {
			raw := streamBody[start:end]
			dec, err := tryFlateDecompress(raw) // tryFlateDecompress from xfdf.go (shared package)
			if err != nil {
				dec, err = tryZlibDecompress(raw)
				if err != nil {
					dec = raw // Fallback
				}
			}
			fullContent.Write(dec)
			fullContent.WriteByte('\n') // Spacing

			// Some PDFs place visible page text inside Form XObjects invoked via Do operators.
			if len(resources) > 0 {
				xobjRefs := resolveUsedXObjectRefs(dec, resources)
				for _, xkey := range xobjRefs {
					appendXObjectContentRecursive(&fullContent, xkey, objMap, visitedXObjects)
				}
			}
		}
	}
	return fullContent.Bytes(), nil
}

func locateStreamSegment(obj []byte) (int, int, bool) {
	streamIdx := bytes.Index(obj, []byte("stream"))
	if streamIdx < 0 {
		return 0, 0, false
	}
	start := streamIdx + len("stream")
	if start < len(obj) && obj[start] == '\r' {
		start++
	}
	if start < len(obj) && obj[start] == '\n' {
		start++
	}

	if l := parseInlineLength(obj); l > 0 && start+l <= len(obj) {
		endByLen := start + l
		k := endByLen
		for k < len(obj) && (obj[k] == '\r' || obj[k] == '\n' || obj[k] == ' ' || obj[k] == '\t') {
			k++
		}
		if k < len(obj) && bytes.HasPrefix(obj[k:], []byte("endstream")) {
			return start, endByLen, true
		}
	}

	endstreamIdx := bytes.Index(obj[start:], []byte("endstream"))
	if endstreamIdx < 0 {
		return 0, 0, false
	}
	end := start + endstreamIdx
	for end > start && (obj[end-1] == '\r' || obj[end-1] == '\n') {
		end--
	}
	if end <= start {
		return 0, 0, false
	}
	return start, end, true
}

func parseInlineLength(obj []byte) int {
	re := regexp.MustCompile(`/Length\s+(\d+)`)
	m := re.FindSubmatch(obj)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(string(m[1]))
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func findPageResources(pageBody []byte, objMap map[string][]byte) []byte {
	if res := extractResourcesBody(pageBody, objMap); len(res) > 0 {
		return res
	}
	parentRe := regexp.MustCompile(`/Parent\s+(\d+)\s+(\d+)\s+R`)
	cur := pageBody
	for depth := 0; depth < 16; depth++ {
		m := parentRe.FindSubmatch(cur)
		if m == nil {
			break
		}
		pkey := string(m[1]) + " " + string(m[2])
		next, ok := objMap[pkey]
		if !ok {
			break
		}
		if res := extractResourcesBody(next, objMap); len(res) > 0 {
			return res
		}
		cur = next
	}
	return nil
}

func extractResourcesBody(body []byte, objMap map[string][]byte) []byte {
	inlineRe := regexp.MustCompile(`(?s)/Resources\s*<<(.*?)>>`)
	if m := inlineRe.FindSubmatch(body); m != nil {
		return m[1]
	}
	refRe := regexp.MustCompile(`/Resources\s+(\d+)\s+(\d+)\s+R`)
	if m := refRe.FindSubmatch(body); m != nil {
		key := string(m[1]) + " " + string(m[2])
		if rb, ok := objMap[key]; ok {
			dictRe := regexp.MustCompile(`(?s)<<(.*?)>>`)
			if dm := dictRe.FindSubmatch(rb); dm != nil {
				return dm[1]
			}
			return rb
		}
	}
	return nil
}

func resolveUsedXObjectRefs(content []byte, resources []byte) []string {
	doRe := regexp.MustCompile(`/([A-Za-z0-9_.+-]+)\s+Do`)
	xobjDictRe := regexp.MustCompile(`(?s)/XObject\s*<<(.*?)>>`)
	m := xobjDictRe.FindSubmatch(resources)
	if m == nil {
		return nil
	}
	xobjDict := m[1]
	nameToRef := map[string]string{}
	refRe := regexp.MustCompile(`/([A-Za-z0-9_.+-]+)\s+(\d+)\s+(\d+)\s+R`)
	for _, r := range refRe.FindAllSubmatch(xobjDict, -1) {
		nameToRef[string(r[1])] = string(r[2]) + " " + string(r[3])
	}
	if len(nameToRef) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, 4)
	for _, d := range doRe.FindAllSubmatch(content, -1) {
		if key, ok := nameToRef[string(d[1])]; ok && !seen[key] {
			seen[key] = true
			out = append(out, key)
		}
	}
	return out
}

func appendXObjectContentRecursive(out *bytes.Buffer, key string, objMap map[string][]byte, visited map[string]bool) {
	if visited[key] {
		return
	}
	visited[key] = true

	body, ok := objMap[key]
	if !ok {
		return
	}
	if !regexp.MustCompile(`/Subtype\s*/Form(\b|\s|/)`).Match(body) {
		return
	}
	raw, dec, ok := inspectStream(body)
	if !ok {
		return
	}
	if len(dec) == 0 {
		dec = raw
	}
	out.Write(dec)
	out.WriteByte('\n')

	res := extractResourcesBody(body, objMap)
	if len(res) == 0 {
		return
	}
	for _, child := range resolveUsedXObjectRefs(dec, res) {
		appendXObjectContentRecursive(out, child, objMap, visited)
	}
}

func isPDFTypePage(body []byte) bool {
	return regexp.MustCompile(`/Type\s*/Page(\b|\s|/)`).FindIndex(body) != nil && !isPDFTypePages(body)
}

func isPDFTypePages(body []byte) bool {
	return regexp.MustCompile(`/Type\s*/Pages(\b|\s|/)`).FindIndex(body) != nil
}

func extractKidsRefs(body []byte) []string {
	refs := make([]string, 0, 4)
	kidsArrRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	if m := kidsArrRe.FindSubmatch(body); m != nil {
		refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
		for _, r := range refRe.FindAllSubmatch(m[1], -1) {
			refs = append(refs, string(r[1])+" "+string(r[2]))
		}
		return refs
	}
	singleKidRe := regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	if m := singleKidRe.FindSubmatch(body); m != nil {
		refs = append(refs, string(m[1])+" "+string(m[2]))
	}
	return refs
}

// estimateStringWidth provides a heuristic-based width estimation based on common
// font character widths. This helps avoid overlaps and gaps caused by assuming
// all characters are the same width.
func estimateStringWidth(text string, fontSize float64) float64 {
	var width float64
	for _, r := range text {
		switch r {
		case 'i', 'j', 'l', 'I', '1', '.', ',', ';', ':', '!', '\'', '|':
			width += 0.25
		case 'f', 't', 'r', '-', ' ', '(', ')':
			width += 0.35
		case 'm', 'w', 'M', 'W', 'O', 'Q', '@', '%':
			width += 0.8
		default:
			switch {
			case r >= 'A' && r <= 'Z':
				width += 0.65
			case r >= '0' && r <= '9':
				width += 0.55
			default:
				width += 0.52
			}
		}
	}
	return width * fontSize
}

func parseTextOperators(content []byte) []models.TextPosition {
	var positions []models.TextPosition

	strContent := string(content)
	btEtRe := regexp.MustCompile(`(?s)BT(.*?)ET`)
	blocks := btEtRe.FindAllStringSubmatch(strContent, -1)
	tmRe := regexp.MustCompile(`([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+Tm`)
	tdRe := regexp.MustCompile(`([\d.-]+)\s+([\d.-]+)\s+T[dD]`)
	tfRe := regexp.MustCompile(`/[A-Za-z0-9_.+-]+\s+([\d.-]+)\s+Tf`)
	textOpRe := regexp.MustCompile(`(?s)\[(?:.|\n|\r)*?\]\s*TJ|<[^>]+>\s*Tj|\((?:\\.|[^\\)])*\)\s*Tj|\((?:\\.|[^\\)])*\)\s*'|[\d.-]+\s+[\d.-]+\s+\((?:\\.|[^\\)])*\)\s*"`)
	tokenRe := regexp.MustCompile(`(?s)[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+Tm|[\d.-]+\s+[\d.-]+\s+T[dD]|/[A-Za-z0-9_.+-]+\s+[\d.-]+\s+Tf|\[(?:.|\n|\r)*?\]\s*TJ|<[^>]+>\s*Tj|\((?:\\.|[^\\)])*\)\s*Tj|\((?:\\.|[^\\)])*\)\s*'|[\d.-]+\s+[\d.-]+\s+\((?:\\.|[^\\)])*\)\s*"`)

	for _, block := range blocks {
		inner := block[1]
		currentX, currentY := 0.0, 0.0
		lineStartX, lineStartY := 0.0, 0.0
		currentFontSize := 10.0
		for _, token := range tokenRe.FindAllString(inner, -1) {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}

			if m := tmRe.FindStringSubmatch(token); m != nil {
				currentX, _ = strconv.ParseFloat(m[5], 64)
				currentY, _ = strconv.ParseFloat(m[6], 64)
				// Tm resets both the current text matrix AND the line matrix.
				lineStartX = currentX
				lineStartY = currentY
				continue
			}
			if m := tdRe.FindStringSubmatch(token); m != nil {
				dx, _ := strconv.ParseFloat(m[1], 64)
				dy, _ := strconv.ParseFloat(m[2], 64)
				// Td/TD moves relative to the line-start matrix (Tlm), not the
				// current text position (which may have advanced after text rendering).
				lineStartX += dx
				lineStartY += dy
				currentX = lineStartX
				currentY = lineStartY
				continue
			}
			if m := tfRe.FindStringSubmatch(token); m != nil {
				if fs, err := strconv.ParseFloat(m[1], 64); err == nil && fs > 0 {
					currentFontSize = fs
				}
				continue
			}
			if !textOpRe.MatchString(token) {
				continue
			}

			textStr := strings.TrimSpace(extractTextFromOperator(token))
			if textStr == "" {
				continue
			}
			height := currentFontSize
			if height < 8 {
				height = 8
			}
			width := estimateStringWidth(textStr, currentFontSize)
			if width < currentFontSize {
				width = currentFontSize
			}
			positions = append(positions, models.TextPosition{
				Text:   textStr,
				X:      currentX,
				Y:      currentY - (0.25 * height),
				Width:  width,
				Height: height,
			})
			// Approximate text advance for subsequent operators on the same text line.
			currentX += width
		}
	}

	return positions
}

func extractTextFromOperator(op string) string {
	op = strings.TrimSpace(op)
	switch {
	case strings.HasSuffix(op, "TJ"):
		start := strings.Index(op, "[")
		end := strings.LastIndex(op, "]")
		if start == -1 || end == -1 || end <= start {
			return ""
		}
		return decodeTJArray(op[start+1 : end])
	case strings.HasSuffix(op, "\""):
		idx := strings.Index(op, "(")
		if idx == -1 {
			return ""
		}
		if lit, ok := readPDFLiteral(op[idx:]); ok {
			return decodePDFLiteral(lit)
		}
	case strings.HasSuffix(op, "Tj") || strings.HasSuffix(op, "'"):
		op = strings.TrimSpace(strings.TrimSuffix(op, "Tj"))
		op = strings.TrimSpace(strings.TrimSuffix(op, "'"))
		if strings.HasPrefix(op, "(") {
			if lit, ok := readPDFLiteral(op); ok {
				return decodePDFLiteral(lit)
			}
		}
		if strings.HasPrefix(op, "<") && strings.HasSuffix(op, ">") {
			return decodePDFHexLiteral(strings.TrimSuffix(strings.TrimPrefix(op, "<"), ">"))
		}
	}
	return ""
}

func decodeTJArray(arr string) string {
	var out strings.Builder
	for i := 0; i < len(arr); {
		switch arr[i] {
		case '(':
			lit, next, ok := readPDFLiteralAt(arr, i)
			if !ok {
				i++
				continue
			}
			out.WriteString(decodePDFLiteral(lit))
			i = next
		case '<':
			j := i + 1
			for j < len(arr) && arr[j] != '>' {
				j++
			}
			if j < len(arr) {
				out.WriteString(decodePDFHexLiteral(arr[i+1 : j]))
				i = j + 1
			} else {
				i++
			}
		default:
			i++
		}
	}
	return out.String()
}

func readPDFLiteral(op string) (string, bool) {
	lit, _, ok := readPDFLiteralAt(op, 0)
	return lit, ok
}

func readPDFLiteralAt(s string, start int) (string, int, bool) {
	if start >= len(s) || s[start] != '(' {
		return "", start, false
	}
	depth := 1
	esc := false
	for i := start + 1; i < len(s); i++ {
		ch := s[i]
		if esc {
			esc = false
			continue
		}
		if ch == '\\' {
			esc = true
			continue
		}
		if ch == '(' {
			depth++
			continue
		}
		if ch == ')' {
			depth--
			if depth == 0 {
				return s[start+1 : i], i + 1, true
			}
		}
	}
	return "", start, false
}

func decodePDFLiteral(s string) string {
	var out bytes.Buffer
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			out.WriteByte(s[i])
			continue
		}
		i++
		if i >= len(s) {
			break
		}
		switch s[i] {
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
			out.WriteByte(s[i])
		case '\n', '\r':
			// line continuation; skip
		default:
			if s[i] >= '0' && s[i] <= '7' {
				val := int(s[i] - '0')
				for k := 0; k < 2 && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '7'; k++ {
					i++
					val = (val * 8) + int(s[i]-'0')
				}
				out.WriteByte(byte(val))
			} else {
				out.WriteByte(s[i])
			}
		}
	}
	return out.String()
}

func decodePDFHexLiteral(hexText string) string {
	hexText = strings.TrimSpace(hexText)
	if hexText == "" {
		return ""
	}
	if len(hexText)%2 != 0 {
		hexText += "0"
	}
	b, err := hex.DecodeString(hexText)
	if err != nil || len(b) == 0 {
		return ""
	}
	if len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF {
		u16 := make([]uint16, 0, (len(b)-2)/2)
		for i := 2; i+1 < len(b); i += 2 {
			u16 = append(u16, (uint16(b[i])<<8)|uint16(b[i+1]))
		}
		return string(utf16.Decode(u16))
	}
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE {
		u16 := make([]uint16, 0, (len(b)-2)/2)
		for i := 2; i+1 < len(b); i += 2 {
			u16 = append(u16, (uint16(b[i+1])<<8)|uint16(b[i]))
		}
		return string(utf16.Decode(u16))
	}
	// Detect UTF-16BE without BOM — standard in CIDFont Identity-H encoded PDFs.
	// Heuristic: even byte count and every high byte (indices 0,2,4…) is 0x00,
	// indicating BMP Unicode code-points encoded as big-endian 16-bit pairs.
	if len(b) >= 4 && len(b)%2 == 0 {
		allHighZero := true
		for i := 0; i < len(b); i += 2 {
			if b[i] != 0x00 {
				allHighZero = false
				break
			}
		}
		if allHighZero {
			u16 := make([]uint16, 0, len(b)/2)
			for i := 0; i+1 < len(b); i += 2 {
				u16 = append(u16, (uint16(b[i])<<8)|uint16(b[i+1]))
			}
			return string(utf16.Decode(u16))
		}
	}
	return string(b)
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
	originalMap, err := buildObjectMap(originalBytes)
	if err != nil {
		return nil, err
	}

	type objMeta struct {
		id  int
		gen int
		key string
	}
	changed := make([]objMeta, 0, 16)
	maxID := 0

	for key, body := range objMap {
		var id, gen int
		if _, scanErr := fmt.Sscanf(key, "%d %d", &id, &gen); scanErr != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}

		origBody, ok := originalMap[key]
		if !ok || !bytes.Equal(origBody, body) {
			changed = append(changed, objMeta{id: id, gen: gen, key: key})
		}
	}

	if len(changed) == 0 {
		return originalBytes, nil
	}

	sort.Slice(changed, func(i, j int) bool {
		if changed[i].id == changed[j].id {
			return changed[i].gen < changed[j].gen
		}
		return changed[i].id < changed[j].id
	})

	prevStartXRef := extractLastStartXRef(originalBytes)
	rootRef, ok := findRootRef(originalBytes)
	if !ok {
		return nil, errors.New("missing Root")
	}
	trailerID := extractPrimaryTrailerID(originalBytes)

	var out bytes.Buffer
	out.Write(originalBytes)
	if len(originalBytes) > 0 {
		last := originalBytes[len(originalBytes)-1]
		if last != '\n' && last != '\r' {
			out.WriteByte('\n')
		}
	}

	offsetByObject := make(map[int]struct {
		offset int
		gen    int
	}, len(changed))

	for _, obj := range changed {
		offsetByObject[obj.id] = struct {
			offset int
			gen    int
		}{offset: out.Len(), gen: obj.gen}

		body := objMap[obj.key]
		fmt.Fprintf(&out, "%d %d obj\n", obj.id, obj.gen)
		out.Write(body)
		if !bytes.HasSuffix(body, []byte("\n")) {
			out.WriteByte('\n')
		}
		if !bytes.HasSuffix(body, []byte("endobj\n")) && !bytes.HasSuffix(body, []byte("endobj")) {
			out.WriteString("endobj\n")
		} else if bytes.HasSuffix(body, []byte("endobj")) {
			out.WriteByte('\n')
		}
	}

	xrefStart := out.Len()
	out.WriteString("xref\n")

	ids := make([]int, 0, len(offsetByObject))
	for id := range offsetByObject {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	start := ids[0]
	block := []int{ids[0]}
	flushBlock := func() {
		if len(block) == 0 {
			return
		}
		out.WriteString(fmt.Sprintf("%d %d\n", start, len(block)))
		for _, id := range block {
			entry := offsetByObject[id]
			out.WriteString(fmt.Sprintf("%010d %05d n \n", entry.offset, entry.gen))
		}
	}

	for i := 1; i < len(ids); i++ {
		if ids[i] == ids[i-1]+1 {
			block = append(block, ids[i])
			continue
		}
		flushBlock()
		start = ids[i]
		block = []int{ids[i]}
	}
	flushBlock()

	trailerIDPart := ""
	if trailerID != "" {
		trailerIDPart = " /ID " + trailerID
	}

	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root %s R /Prev %d%s >>\nstartxref\n%d\n%%%%EOF\n", maxID+1, rootRef, prevStartXRef, trailerIDPart, xrefStart)

	return out.Bytes(), nil
}

func extractPrimaryTrailerID(pdfBytes []byte) string {
	if len(pdfBytes) == 0 {
		return ""
	}
	trailerRe := regexp.MustCompile(`(?s)trailer\s*<<(.*?)>>`)
	idRe := regexp.MustCompile(`(?s)/ID\s*(\[(?:.|\n|\r)*?\])`)

	if tm := trailerRe.FindSubmatch(pdfBytes); tm != nil {
		if idm := idRe.FindSubmatch(tm[1]); idm != nil {
			return strings.TrimSpace(string(idm[1]))
		}
	}

	// Fallback for PDFs with non-standard trailer layout.
	if idm := idRe.FindSubmatch(pdfBytes); idm != nil {
		return strings.TrimSpace(string(idm[1]))
	}

	return ""
}

func extractLastStartXRef(pdfBytes []byte) int {
	if len(pdfBytes) == 0 {
		return 0
	}
	re := regexp.MustCompile(`(?s)startxref\s*(\d+)\s*%%EOF\s*$`)
	if m := re.FindSubmatch(pdfBytes); m != nil {
		if n, err := strconv.Atoi(string(m[1])); err == nil {
			return n
		}
	}
	reAny := regexp.MustCompile(`startxref\s*(\d+)`)
	all := reAny.FindAllSubmatch(pdfBytes, -1)
	if len(all) == 0 {
		return 0
	}
	last := all[len(all)-1]
	if len(last) < 2 {
		return 0
	}
	n, err := strconv.Atoi(string(last[1]))
	if err != nil {
		return 0
	}
	return n
}
