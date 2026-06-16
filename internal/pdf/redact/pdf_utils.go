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

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

var (
	btEtRe   = regexp.MustCompile(`(?s)BT(.*?)ET`)
	tmRe     = regexp.MustCompile(`([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+Tm`)
	tdRe     = regexp.MustCompile(`([\d.-]+)\s+([\d.-]+)\s+T[dD]`)
	tfRe     = regexp.MustCompile(`/[A-Za-z0-9_.+-]+\s+([\d.-]+)\s+Tf`)
	textOpRe = regexp.MustCompile(`(?s)\[(?:.|\n|\r)*?\]\s*TJ|<[^>]+>\s*Tj|\((?:\\.|[^\\)])*\)\s*Tj|\((?:\\.|[^\\)])*\)\s*'|[\d.-]+\s+[\d.-]+\s+\((?:\\.|[^\\)])*\)\s*"`)
	tokenRe  = regexp.MustCompile(`(?s)[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+[\d.-]+\s+Tm|[\d.-]+\s+[\d.-]+\s+T[dD]|/[A-Za-z0-9_.+-]+\s+[\d.-]+\s+Tf|\[(?:.|\n|\r)*?\]\s*TJ|<[^>]+>\s*Tj|\((?:\\.|[^\\)])*\)\s*Tj|\((?:\\.|[^\\)])*\)\s*'|[\d.-]+\s+[\d.-]+\s+\((?:\\.|[^\\)])*\)\s*"`)
)

// buildObjectMap parses the PDF into object-number → body slices and complements
// it with xref-stream and object-stream expansion via the merge helpers.
func buildObjectMap(pdfBytes []byte) (map[int][]byte, map[int]int, error) {
	objMap := make(map[int][]byte)
	objGen := make(map[int]int)

	for _, b := range merge.FindObjectBoundaries(pdfBytes) {
		bodyEnd := b.End - len("endobj")
		for bodyEnd > b.BodyStart && isPDFWhitespace(pdfBytes[bodyEnd-1]) {
			bodyEnd--
		}
		body := pdfBytes[b.BodyStart:bodyEnd]
		objMap[b.ObjNum] = body
		objGen[b.ObjNum] = b.GenNum

		if merge.IsObjectStream(body) {
			for onum, frag := range merge.ParseObjectStream(body) {
				objMap[onum] = frag
				objGen[onum] = 0
			}
		}
	}

	parseXRefStreams(pdfBytes, objMap, objGen)
	return objMap, objGen, nil
}

func traversePages(pageTreeObjNum int, objMap map[int][]byte, dims *[]models.PageDetail) error {
	body, ok := objMap[pageTreeObjNum]
	if !ok {
		return nil
	}

	if isPDFTypePages(body) {
		for _, kidNum := range extractKidsRefs(body) {
			if err := traversePages(kidNum, objMap, dims); err != nil {
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

func extractMediaBox(body []byte, objMap map[int][]byte) [4]float64 {
	defaultBox := [4]float64{0, 0, 595.28, 841.89}
	rectRe := regexp.MustCompile(`/MediaBox\s*\[\s*([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s*\]`)
	match := rectRe.FindSubmatch(body)

	if match == nil {
		refRe := regexp.MustCompile(`/MediaBox\s+(\d+)\s+(\d+)\s+R`)
		refMatch := refRe.FindSubmatch(body)
		if refMatch != nil {
			refNum, _ := strconv.Atoi(string(refMatch[1]))
			if refBody, ok := objMap[refNum]; ok {
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

func findPageObject(objMap map[int][]byte, pdfBytes []byte, targetPage int) (int, error) {
	rootNum, _, ok := findRootRef(pdfBytes)
	if !ok {
		return 0, errors.New("missing Root")
	}

	rootBody := objMap[rootNum]
	pagesRefRe := regexp.MustCompile(`/Pages\s+(\d+)\s+(\d+)\s+R`)
	pm := pagesRefRe.FindSubmatch(rootBody)
	if pm == nil {
		return 0, errors.New("missing /Pages in Root")
	}
	pagesNum, err := strconv.Atoi(string(pm[1]))
	if err != nil || pagesNum <= 0 {
		return 0, errors.New("invalid /Pages in Root")
	}

	var foundKey int
	found := false
	var currentPage int

	var walk func(int) error
	walk = func(objNum int) error {
		if found {
			return nil
		}
		body, ok := objMap[objNum]
		if !ok {
			return nil
		}

		if isPDFTypePages(body) {
			for _, kidNum := range extractKidsRefs(body) {
				if err := walk(kidNum); err != nil {
					return err
				}
			}
			return nil
		}

		if isPDFTypePage(body) {
			currentPage++
			if currentPage == targetPage {
				foundKey = objNum
				found = true
			}
			return nil
		}
		return nil
	}

	if err := walk(pagesNum); err != nil {
		return 0, err
	}

	if !found {
		return 0, fmt.Errorf("page %d not found", targetPage)
	}
	return foundKey, nil
}

func extractPageContent(pageBody []byte, objMap map[int][]byte) ([]byte, error) {
	contentsRe := regexp.MustCompile(`/Contents\s+(?:(\d+)\s+(\d+)\s+R|\[(.*?)\])`)
	match := contentsRe.FindSubmatch(pageBody)
	if match == nil {
		return nil, nil // Empty content
	}
	resources := findPageResources(pageBody, objMap)

	var contentNums []int
	if len(match[1]) > 0 {
		n, err := strconv.Atoi(string(match[1]))
		if err == nil {
			contentNums = append(contentNums, n)
		}
	} else if len(match[3]) > 0 {
		refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
		refs := refRe.FindAllSubmatch(match[3], -1)
		for _, r := range refs {
			n, err := strconv.Atoi(string(r[1]))
			if err == nil {
				contentNums = append(contentNums, n)
			}
		}
	}

	var fullContent bytes.Buffer
	visitedXObjects := map[int]bool{}
	for _, objNum := range contentNums {
		streamBody, ok := objMap[objNum]
		if !ok {
			continue
		}
		start, end, ok := locateStreamSegment(streamBody)
		if ok {
			raw := streamBody[start:end]
			dec, err := tryFlateDecompress(raw)
			if err != nil {
				dec, err = tryZlibDecompress(raw)
				if err != nil {
					dec = raw // Fallback
				}
			}
			fullContent.Write(dec)
			fullContent.WriteByte('\n') // Spacing

			if len(resources) > 0 {
				xobjRefs := resolveUsedXObjectRefs(dec, resources)
				for _, xnum := range xobjRefs {
					appendXObjectContentRecursive(&fullContent, xnum, objMap, visitedXObjects)
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

	endstreamIdx := bytes.LastIndex(obj, []byte("endstream"))
	if endstreamIdx < 0 || endstreamIdx < start {
		return 0, 0, false
	}
	end := endstreamIdx
	for end > start && (obj[end-1] == '\r' || obj[end-1] == '\n') {
		end--
	}
	if end <= start {
		return 0, 0, false
	}
	return start, end, true
}

func parseInlineLength(obj []byte) int {
	if regexp.MustCompile(`/Length\s+\d+\s+\d+\s+R`).Find(obj) != nil {
		return 0
	}

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

func findPageResources(pageBody []byte, objMap map[int][]byte) []byte {
	if res := extractResourcesBody(pageBody, objMap); len(res) > 0 {
		return res
	}
	parentRe := regexp.MustCompile(`/Parent\s+(\d+)\s+(\d+)\s+R`)
	cur := pageBody
	for range 16 {
		m := parentRe.FindSubmatch(cur)
		if m == nil {
			break
		}
		parentNum, err := strconv.Atoi(string(m[1]))
		if err != nil || parentNum <= 0 {
			break
		}
		next, ok := objMap[parentNum]
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

func extractResourcesBody(body []byte, objMap map[int][]byte) []byte {
	inlineRe := regexp.MustCompile(`(?s)/Resources\s*<<(.*?)>>`)
	if m := inlineRe.FindSubmatch(body); m != nil {
		return m[1]
	}
	refRe := regexp.MustCompile(`/Resources\s+(\d+)\s+(\d+)\s+R`)
	if m := refRe.FindSubmatch(body); m != nil {
		resNum, err := strconv.Atoi(string(m[1]))
		if err != nil {
			return nil
		}
		if rb, ok := objMap[resNum]; ok {
			dictRe := regexp.MustCompile(`(?s)<<(.*?)>>`)
			if dm := dictRe.FindSubmatch(rb); dm != nil {
				return dm[1]
			}
			return rb
		}
	}
	return nil
}

func resolveUsedXObjectRefs(content []byte, resources []byte) []int {
	doRe := regexp.MustCompile(`/([A-Za-z0-9_.+-]+)\s+Do`)
	xobjDictRe := regexp.MustCompile(`(?s)/XObject\s*<<(.*?)>>`)
	m := xobjDictRe.FindSubmatch(resources)
	if m == nil {
		return nil
	}
	xobjDict := m[1]
	nameToRef := map[string]int{}
	refRe := regexp.MustCompile(`/([A-Za-z0-9_.+-]+)\s+(\d+)\s+(\d+)\s+R`)
	for _, r := range refRe.FindAllSubmatch(xobjDict, -1) {
		num, err := strconv.Atoi(string(r[2]))
		if err == nil {
			nameToRef[string(r[1])] = num
		}
	}
	if len(nameToRef) == 0 {
		return nil
	}
	seen := map[int]bool{}
	out := make([]int, 0, 4)
	for _, d := range doRe.FindAllSubmatch(content, -1) {
		if num, ok := nameToRef[string(d[1])]; ok && !seen[num] {
			seen[num] = true
			out = append(out, num)
		}
	}
	return out
}

func appendXObjectContentRecursive(out *bytes.Buffer, objNum int, objMap map[int][]byte, visited map[int]bool) {
	if visited[objNum] {
		return
	}
	visited[objNum] = true

	body, ok := objMap[objNum]
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

func extractKidsRefs(body []byte) []int {
	refs := make([]int, 0, 4)
	kidsArrRe := regexp.MustCompile(`/Kids\s*\[(.*?)\]`)
	if m := kidsArrRe.FindSubmatch(body); m != nil {
		refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
		for _, r := range refRe.FindAllSubmatch(m[1], -1) {
			n, err := strconv.Atoi(string(r[1]))
			if err == nil {
				refs = append(refs, n)
			}
		}
		return refs
	}
	singleKidRe := regexp.MustCompile(`/Kids\s+(\d+)\s+(\d+)\s+R`)
	if m := singleKidRe.FindSubmatch(body); m != nil {
		n, err := strconv.Atoi(string(m[1]))
		if err == nil {
			refs = append(refs, n)
		}
	}
	return refs
}

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
	blocks := btEtRe.FindAllStringSubmatch(strContent, -1)

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
				lineStartX = currentX
				lineStartY = currentY
				continue
			}
			if m := tdRe.FindStringSubmatch(token); m != nil {
				dx, _ := strconv.ParseFloat(m[1], 64)
				dy, _ := strconv.ParseFloat(m[2], 64)
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

func appendStreamToPage(pageBody []byte, streamObjNum, streamGenNum int) []byte {
	contentsRe := regexp.MustCompile(`/Contents\s+(?:(\d+\s+\d+\s+R)|\[(.*?)\])`)
	match := contentsRe.FindSubmatchIndex(pageBody)

	refStr := fmt.Sprintf("%d %d R", streamObjNum, streamGenNum)

	if match == nil {
		dictEnd := bytes.LastIndex(pageBody, []byte(">>"))
		if dictEnd == -1 {
			return pageBody
		}
		var buf bytes.Buffer
		buf.Write(pageBody[:dictEnd])
		buf.WriteString(fmt.Sprintf("/Contents [%s]", refStr))
		buf.Write(pageBody[dictEnd:])
		return buf.Bytes()
	}

	start, end := match[0], match[1]

	var replacement string
	if match[2] != -1 && match[3] != -1 {
		oldRef := string(pageBody[match[2]:match[3]])
		replacement = fmt.Sprintf("/Contents [%s %s]", oldRef, refStr)
	} else if match[4] != -1 && match[5] != -1 {
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

func rebuildPDF(objMap map[int][]byte, objGen map[int]int, originalBytes []byte) ([]byte, error) {
	originalMap, _, err := buildObjectMap(originalBytes)
	if err != nil {
		return nil, err
	}

	type objMeta struct {
		id  int
		gen int
	}
	changed := make([]objMeta, 0, 16)
	maxID := 0
	for objNum, body := range objMap {
		if objNum > maxID {
			maxID = objNum
		}
		gen := objGenNum(objGen, objNum)
		origBody, ok := originalMap[objNum]
		if !ok || len(origBody) != len(body) || !bytes.Equal(origBody, body) {
			changed = append(changed, objMeta{id: objNum, gen: gen})
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
	rootNum, rootGen, ok := findRootRef(originalBytes)
	if !ok {
		return nil, errors.New("missing Root")
	}
	rootRefStr := fmt.Sprintf("%d %d", rootNum, rootGen)
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

	var objHeader [32]byte
	for _, obj := range changed {
		offsetByObject[obj.id] = struct {
			offset int
			gen    int
		}{offset: out.Len(), gen: obj.gen}

		body := objMap[obj.id]
		header := strconv.AppendInt(objHeader[:0], int64(obj.id), 10)
		header = append(header, ' ')
		header = strconv.AppendInt(header, int64(obj.gen), 10)
		header = append(header, ' ', 'o', 'b', 'j', '\n')
		out.Write(header)
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
			out.Write(xrefEntryBytes(entry.offset, entry.gen))
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

	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root %s R /Prev %d%s >>\nstartxref\n%d\n%%%%EOF\n", maxID+1, rootRefStr, prevStartXRef, trailerIDPart, xrefStart)

	return out.Bytes(), nil
}

func xrefEntryBytes(offset, gen int) []byte {
	var buf [20]byte
	pos := 9
	off := offset
	for range 10 {
		buf[pos] = byte('0' + off%10)
		off /= 10
		pos--
	}
	buf[10] = ' '
	pos = 15
	g := gen
	for range 5 {
		buf[pos] = byte('0' + g%10)
		g /= 10
		pos--
	}
	buf[16] = ' '
	buf[17] = 'n'
	buf[18] = ' '
	buf[19] = '\n'
	return buf[:]
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
