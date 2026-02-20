package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
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

// RedactionTextQuery describes text-based redaction criteria.
type RedactionTextQuery struct {
	Text string `json:"text"`
}

// ApplyRedactionOptions represents a unified redaction request.
type ApplyRedactionOptions struct {
	Blocks     []RedactionRect      `json:"blocks,omitempty"`
	TextSearch []RedactionTextQuery `json:"textSearch,omitempty"`
	Mode       string               `json:"mode,omitempty"`     // secure_required | visual_allowed
	Password   string               `json:"password,omitempty"` // reserved for encrypted inputs
	OCR        *OCRSettings         `json:"ocr,omitempty"`
}

// OCRSettings is an extension point for OCR providers.
type OCRSettings struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider,omitempty"`
	Language string `json:"language,omitempty"`
}

// PageCapability describes whether a page contains text or image-like content.
type PageCapability struct {
	PageNum   int    `json:"pageNum"`
	Type      string `json:"type"` // text | image_only | mixed | unknown
	HasText   bool   `json:"hasText"`
	HasImage  bool   `json:"hasImage"`
	OCREnable bool   `json:"ocrEnabled"`
	Note      string `json:"note,omitempty"`
}

// RedactionApplyReport provides explicit safety/capability metadata.
type RedactionApplyReport struct {
	Mode              string           `json:"mode"`
	SecurityOutcome   string           `json:"securityOutcome"` // secure|visual_only|failed
	AppliedSecure     bool             `json:"appliedSecure"`
	AppliedVisual     bool             `json:"appliedVisual"`
	GeneratedRects    int              `json:"generatedRects"`
	AppliedRectangles int              `json:"appliedRectangles"`
	MatchedTextCount  int              `json:"matchedTextCount"`
	Capabilities      []PageCapability `json:"capabilities,omitempty"`
	UnsupportedPages  []int            `json:"unsupportedPages,omitempty"`
	Warnings          []string         `json:"warnings,omitempty"`
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
	normalizedQuery := normalizeSearchText(searchText)
	searchText = strings.ToLower(searchText)

	for i := 1; i <= info.TotalPages; i++ {
		positions, err := ExtractTextPositions(pdfBytes, i)
		if err != nil {
			// Log error but continue? Or fail? Best to continue for search flexibility
			continue
		}

		for _, pos := range positions {
			redactions = append(redactions, buildSubstringRects(i, pos, searchText)...)
		}

		// Fallback for PDFs that split phrases across multiple text-show operators.
		if len(positions) > 1 && strings.Contains(normalizedQuery, " ") {
			if rect, ok := findCombinedMatchRect(i, positions, normalizedQuery); ok {
				redactions = append(redactions, rect)
			}
		}
	}
	return redactions, nil
}

func buildSubstringRects(pageNum int, pos TextPosition, loweredSearch string) []RedactionRect {
	if loweredSearch == "" || strings.TrimSpace(pos.Text) == "" {
		return nil
	}
	src := []rune(strings.ToLower(pos.Text))
	needle := []rune(loweredSearch)
	if len(src) == 0 || len(needle) == 0 || len(needle) > len(src) {
		return nil
	}

	charW := 0.0
	if pos.Width > 0 {
		charW = pos.Width / float64(len(src))
	}

	rects := make([]RedactionRect, 0, 2)
	for i := 0; i+len(needle) <= len(src); i++ {
		if !runeSliceEqual(src[i:i+len(needle)], needle) {
			continue
		}
		x := pos.X
		w := pos.Width
		if charW > 0 {
			x = pos.X + (float64(i) * charW)
			w = float64(len(needle)) * charW
		}
		rects = append(rects, RedactionRect{
			PageNum: pageNum,
			X:       x,
			Y:       pos.Y,
			Width:   w,
			Height:  pos.Height,
		})
	}
	return rects
}

func runeSliceEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func normalizeSearchText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

func findCombinedMatchRect(pageNum int, positions []TextPosition, normalizedQuery string) (RedactionRect, bool) {
	ordered := append([]TextPosition(nil), positions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Y == ordered[j].Y {
			return ordered[i].X < ordered[j].X
		}
		return ordered[i].Y > ordered[j].Y
	})

	for i := 0; i < len(ordered); i++ {
		joined := ""
		minX := ordered[i].X
		minY := ordered[i].Y
		maxX := ordered[i].X + ordered[i].Width
		maxY := ordered[i].Y + ordered[i].Height

		for j := i; j < len(ordered) && j < i+24; j++ {
			part := strings.TrimSpace(ordered[j].Text)
			if part == "" {
				continue
			}
			if joined == "" {
				joined = part
			} else {
				joined += " " + part
			}

			if ordered[j].X < minX {
				minX = ordered[j].X
			}
			if ordered[j].Y < minY {
				minY = ordered[j].Y
			}
			if x := ordered[j].X + ordered[j].Width; x > maxX {
				maxX = x
			}
			if y := ordered[j].Y + ordered[j].Height; y > maxY {
				maxY = y
			}

			if strings.Contains(normalizeSearchText(joined), normalizedQuery) {
				return RedactionRect{
					PageNum: pageNum,
					X:       minX,
					Y:       minY,
					Width:   maxX - minX,
					Height:  maxY - minY,
				}, true
			}
		}
	}

	return RedactionRect{}, false
}

// ApplyRedactionsAdvanced applies a unified redaction request.
// Current implementation supports visual overlay redactions only.
func ApplyRedactionsAdvanced(pdfBytes []byte, opts ApplyRedactionOptions) ([]byte, error) {
	out, _, err := ApplyRedactionsAdvancedWithReport(pdfBytes, opts)
	return out, err
}

// ApplyRedactionsAdvancedWithReport applies redactions and returns an execution report.
func ApplyRedactionsAdvancedWithReport(pdfBytes []byte, opts ApplyRedactionOptions) ([]byte, RedactionApplyReport, error) {
	if len(pdfBytes) == 0 {
		return nil, RedactionApplyReport{}, errors.New("empty pdf bytes")
	}

	report := RedactionApplyReport{
		Mode:            "visual_allowed",
		SecurityOutcome: "visual_only",
	}

	mode := strings.TrimSpace(strings.ToLower(opts.Mode))
	if mode == "" {
		mode = "visual_allowed"
	}
	report.Mode = mode
	if mode != "visual_allowed" && mode != "secure_required" {
		return nil, report, errors.New("invalid mode: expected visual_allowed or secure_required")
	}

	if trailerHasEncrypt(pdfBytes) {
		dec, err := decryptEncryptedPDFBytes(pdfBytes, opts.Password)
		if err != nil {
			return nil, report, err
		}
		pdfBytes = dec
		report.Warnings = append(report.Warnings, "input PDF was decrypted using in-house pipeline and output is emitted decrypted")
	}

	caps, capErr := AnalyzePageCapabilities(pdfBytes)
	if capErr == nil {
		report.Capabilities = caps
	}

	if opts.OCR != nil && opts.OCR.Enabled {
		report.Warnings = append(report.Warnings, "OCR requested but no OCR provider is configured; this is an extension hook")
	}

	all := make([]RedactionRect, 0, len(opts.Blocks)+8)
	all = append(all, opts.Blocks...)

	for _, q := range opts.TextSearch {
		query := strings.TrimSpace(q.Text)
		if query == "" {
			continue
		}
		rects, err := FindTextOccurrences(pdfBytes, query)
		if err != nil {
			return nil, report, err
		}
		all = append(all, rects...)
		report.MatchedTextCount += len(rects)
	}

	if opts.OCR != nil && opts.OCR.Enabled {
		ocrRects, err := runOCRSearch(pdfBytes, opts.TextSearch, *opts.OCR)
		if err != nil {
			report.Warnings = append(report.Warnings, "OCR fallback error: "+err.Error())
		} else {
			all = append(all, ocrRects...)
			report.MatchedTextCount += len(ocrRects)
		}
	}
	report.GeneratedRects = len(all)
	workingPDF := pdfBytes

	// In visual mode, still attempt secure content rewriting as best-effort.
	// If it works, we get both searchable-text removal and visual overlays.
	if mode == "visual_allowed" && len(all) > 0 {
		secureOut, secureChanged, secureWarns, err := applySecureContentRedactions(workingPDF, all, opts.TextSearch)
		report.Warnings = append(report.Warnings, secureWarns...)
		if err != nil {
			report.Warnings = append(report.Warnings, "best-effort secure removal skipped: "+err.Error())
		} else if secureChanged {
			workingPDF = secureOut
			report.AppliedSecure = true
			report.SecurityOutcome = "secure"
		}
	}

	if mode == "secure_required" {
		secureOut, secureChanged, secureWarns, err := applySecureContentRedactions(workingPDF, all, opts.TextSearch)
		report.Warnings = append(report.Warnings, secureWarns...)
		if err != nil {
			report.SecurityOutcome = "failed"
			return nil, report, err
		}
		if !secureChanged {
			report.SecurityOutcome = "failed"
			return nil, report, errors.New("secure_required requested but no secure text content could be removed")
		}
		visualOut, err := ApplyRedactions(secureOut, all)
		if err != nil {
			report.SecurityOutcome = "failed"
			return nil, report, err
		}
		report.AppliedSecure = true
		report.AppliedVisual = true
		report.SecurityOutcome = "secure"
		report.AppliedRectangles = len(all)
		return visualOut, report, nil
	}

	out, err := ApplyRedactions(workingPDF, all)
	if err != nil {
		report.SecurityOutcome = "failed"
		return nil, report, err
	}
	report.AppliedVisual = true
	report.AppliedRectangles = len(all)
	return out, report, nil
}

// AnalyzePageCapabilities classifies each page for text/image redaction capability.
func AnalyzePageCapabilities(pdfBytes []byte) ([]PageCapability, error) {
	objMap, err := buildObjectMap(pdfBytes)
	if err != nil {
		return nil, err
	}
	info, err := GetPageInfo(pdfBytes)
	if err != nil {
		return nil, err
	}
	caps := make([]PageCapability, 0, info.TotalPages)
	for i := 1; i <= info.TotalPages; i++ {
		pageRef, err := findPageObject(objMap, pdfBytes, i)
		if err != nil {
			caps = append(caps, PageCapability{PageNum: i, Type: "unknown", Note: err.Error()})
			continue
		}
		body := objMap[pageRef]
		keys := extractContentKeys(body)
		hasText := false
		hasImage := false
		for _, key := range keys {
			objBody, ok := objMap[key]
			if !ok {
				continue
			}
			rawStream, decStream, _ := inspectStream(objBody)
			combined := append(rawStream, decStream...)
			s := string(combined)
			if strings.Contains(s, "BT") && (strings.Contains(s, "Tj") || strings.Contains(s, "TJ")) {
				hasText = true
			}
			if strings.Contains(s, " Do") || bytesIndex(objBody, []byte("/Image")) >= 0 {
				hasImage = true
			}
		}
		if len(keys) == 0 {
			content, _ := extractPageContent(body, objMap)
			s := string(content)
			hasText = strings.Contains(s, "BT") && (strings.Contains(s, "Tj") || strings.Contains(s, "TJ"))
			hasImage = strings.Contains(s, " Do") || bytesIndex(body, []byte("/Image")) >= 0
		}
		typeName := "unknown"
		switch {
		case hasText && hasImage:
			typeName = "mixed"
		case hasText:
			typeName = "text"
		case hasImage:
			typeName = "image_only"
		}
		cap := PageCapability{PageNum: i, Type: typeName, HasText: hasText, HasImage: hasImage}
		if typeName == "image_only" {
			cap.Note = "text search requires OCR for image-only content"
		}
		caps = append(caps, cap)
	}
	return caps, nil
}

func applySecureContentRedactions(pdfBytes []byte, redactions []RedactionRect, queries []RedactionTextQuery) ([]byte, bool, []string, error) {
	objMap, err := buildObjectMap(pdfBytes)
	if err != nil {
		return nil, false, nil, err
	}

	redactionsByPage := make(map[int][]RedactionRect)
	for _, r := range redactions {
		redactionsByPage[r.PageNum] = append(redactionsByPage[r.PageNum], r)
	}
	if len(redactionsByPage) == 0 && len(queries) > 0 {
		if info, err := GetPageInfo(pdfBytes); err == nil {
			for i := 1; i <= info.TotalPages; i++ {
				redactionsByPage[i] = nil
			}
		}
	}

	var warnings []string
	changedAny := false

	for pageNum, rects := range redactionsByPage {
		pageRef, err := findPageObject(objMap, pdfBytes, pageNum)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("page %d: %v", pageNum, err))
			continue
		}
		pageBody := objMap[pageRef]
		keys := extractContentKeys(pageBody)
		if len(keys) == 0 {
			warnings = append(warnings, fmt.Sprintf("page %d: no content streams", pageNum))
			continue
		}

		for _, key := range keys {
			objBody, ok := objMap[key]
			if !ok {
				continue
			}
			updated, changed, err := rewriteContentStreamSecure(objBody, rects, queries)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("stream %s: %v", key, err))
				continue
			}
			if changed {
				objMap[key] = updated
				changedAny = true
			}
		}
	}

	out, err := rebuildPDF(objMap, pdfBytes)
	if err != nil {
		return nil, false, warnings, err
	}

	return out, changedAny, warnings, nil
}

func inspectStream(streamObj []byte) ([]byte, []byte, bool) {
	streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	loc := streamRe.FindSubmatchIndex(streamObj)
	if loc == nil {
		return nil, nil, false
	}
	raw := streamObj[loc[2]:loc[3]]
	if bytesIndex(streamObj, []byte("/FlateDecode")) >= 0 {
		if d, err := tryFlateDecompress(raw); err == nil {
			return raw, d, true
		}
		if d, err := tryZlibDecompress(raw); err == nil {
			return raw, d, true
		}
	}
	return raw, raw, true
}

func extractContentKeys(pageBody []byte) []string {
	contentsRe := regexp.MustCompile(`/Contents\s+(?:(\d+)\s+(\d+)\s+R|\[(.*?)\])`)
	match := contentsRe.FindSubmatch(pageBody)
	if match == nil {
		return nil
	}
	var keys []string
	if len(match[1]) > 0 {
		keys = append(keys, string(match[1])+" "+string(match[2]))
		return keys
	}
	if len(match[3]) > 0 {
		refRe := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
		refs := refRe.FindAllSubmatch(match[3], -1)
		for _, r := range refs {
			keys = append(keys, string(r[1])+" "+string(r[2]))
		}
	}
	return keys
}

func rewriteContentStreamSecure(streamObj []byte, rects []RedactionRect, queries []RedactionTextQuery) ([]byte, bool, error) {
	streamRe := regexp.MustCompile(`(?s)stream\s*\r?\n(.*?)\r?\nendstream`)
	loc := streamRe.FindSubmatchIndex(streamObj)
	if loc == nil {
		return streamObj, false, nil
	}
	raw := streamObj[loc[2]:loc[3]]

	decoded := raw
	compressed := false
	if bytesIndex(streamObj, []byte("/FlateDecode")) >= 0 {
		if d, err := tryFlateDecompress(raw); err == nil {
			decoded = d
			compressed = true
		} else if d, err := tryZlibDecompress(raw); err == nil {
			decoded = d
			compressed = true
		} else {
			return streamObj, false, errors.New("unable to decode flate stream")
		}
	}

	newDecoded, changed := scrubDecodedContent(decoded, rects, queries)
	if !changed {
		return streamObj, false, nil
	}

	encoded := newDecoded
	if compressed {
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)
		if _, err := zw.Write(newDecoded); err != nil {
			return nil, false, err
		}
		if err := zw.Close(); err != nil {
			return nil, false, err
		}
		encoded = buf.Bytes()
	}

	newObj := make([]byte, 0, len(streamObj)+64)
	newObj = append(newObj, streamObj[:loc[2]]...)
	newObj = append(newObj, encoded...)
	newObj = append(newObj, streamObj[loc[3]:]...)

	lenRe := regexp.MustCompile(`/Length\s+\d+`)
	newObj = lenRe.ReplaceAll(newObj, []byte(fmt.Sprintf("/Length %d", len(encoded))))

	return newObj, true, nil
}

func scrubDecodedContent(decoded []byte, rects []RedactionRect, queries []RedactionTextQuery) ([]byte, bool) {
	positions := parseTextOperators(decoded)

	opRe := regexp.MustCompile(`(?s)\[(?:.|\n|\r)*?\]\s*TJ|<[^>]+>\s*Tj|\((?:\\.|[^\\)])*\)\s*Tj|\((?:\\.|[^\\)])*\)\s*'|[\d.-]+\s+[\d.-]+\s+\((?:\\.|[^\\)])*\)\s*"`)
	src := string(decoded)
	matches := opRe.FindAllStringIndex(src, -1)
	if len(matches) == 0 {
		return decoded, false
	}

	var out strings.Builder
	last := 0
	changed := false
	posIdx := 0

	for _, m := range matches {
		out.WriteString(src[last:m[0]])
		op := src[m[0]:m[1]]
		text := strings.TrimSpace(extractTextFromOperator(op))
		if text == "" {
			out.WriteString(op)
			last = m[1]
			continue
		}

		newText := text
		if posIdx < len(positions) {
			p := positions[posIdx]
			if strings.TrimSpace(p.Text) != text {
				for lookahead := posIdx + 1; lookahead < len(positions) && lookahead < posIdx+6; lookahead++ {
					if strings.TrimSpace(positions[lookahead].Text) == text {
						p = positions[lookahead]
						posIdx = lookahead
						break
					}
				}
			}
			newText = applyRectMaskToText(newText, p, rects)
			posIdx++
		}

		for _, q := range queries {
			term := strings.TrimSpace(q.Text)
			if term == "" {
				continue
			}
			newText = replaceCaseInsensitiveWithSpaces(newText, term)
		}

		if newText != text {
			changed = true
			out.WriteString("(")
			out.WriteString(escapePDFTextLiteral(newText))
			out.WriteString(") Tj")
		} else {
			out.WriteString(op)
		}
		last = m[1]
	}

	out.WriteString(src[last:])
	if !changed {
		return decoded, false
	}
	return []byte(out.String()), true
}

func applyRectMaskToText(text string, pos TextPosition, rects []RedactionRect) string {
	runes := []rune(text)
	if len(runes) == 0 || pos.Width <= 0 {
		return text
	}
	charW := pos.Width / float64(len(runes))
	if charW <= 0 {
		return text
	}
	out := append([]rune(nil), runes...)
	for _, r := range rects {
		if !rectsIntersect(pos.X, pos.Y, pos.Width, pos.Height, r.X, r.Y, r.Width, r.Height) {
			continue
		}
		start := int(math.Round((r.X - pos.X) / charW))
		end := int(math.Round((r.X + r.Width - pos.X) / charW))
		if end <= start {
			end = start + 1
		}
		if start < 0 {
			start = 0
		}
		if end > len(out) {
			end = len(out)
		}
		if start >= end {
			continue
		}
		for i := start; i < end; i++ {
			out[i] = ' '
		}
	}
	return string(out)
}

func replaceCaseInsensitiveWithSpaces(s, term string) string {
	if term == "" {
		return s
	}
	lower := strings.ToLower(s)
	target := strings.ToLower(term)
	if !strings.Contains(lower, target) {
		return s
	}
	b := []rune(s)
	lr := []rune(lower)
	tr := []rune(target)
	for i := 0; i+len(tr) <= len(lr); {
		if string(lr[i:i+len(tr)]) == string(tr) {
			for j := i; j < i+len(tr); j++ {
				b[j] = ' '
				lr[j] = ' '
			}
			i += len(tr)
			continue
		}
		i++
	}
	return string(b)
}

func escapePDFTextLiteral(s string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return replacer.Replace(s)
}

func rectsIntersect(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
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

func traversePages(key string, objMap map[string][]byte, dims *[]PageDetail) error {
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
		*dims = append(*dims, PageDetail{
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

func parseTextOperators(content []byte) []TextPosition {
	var positions []TextPosition

	strContent := string(content)
	btEtRe := regexp.MustCompile(`(?s)BT(.*?)ET`)
	blocks := btEtRe.FindAllStringSubmatch(strContent, -1)
	tmRe := regexp.MustCompile(`([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+Tm`)
	tdRe := regexp.MustCompile(`([\d.-]+)\s+([\d.-]+)\s+Td`)
	tfRe := regexp.MustCompile(`/[A-Za-z0-9_.+-]+\s+([\d.-]+)\s+Tf`)
	opRe := regexp.MustCompile(`(?s)\[(?:.|\n|\r)*?\]\s*TJ|<[^>]+>\s*Tj|\((?:\\.|[^\\)])*\)\s*Tj|\((?:\\.|[^\\)])*\)\s*'|[\d.-]+\s+[\d.-]+\s+\((?:\\.|[^\\)])*\)\s*"`)

	for _, block := range blocks {
		inner := block[1]
		currentX, currentY := 0.0, 0.0
		currentFontSize := 10.0
		for _, m := range opRe.FindAllStringIndex(inner, -1) {
			op := inner[m[0]:m[1]]
			prefix := inner[:m[0]]

			tmMatches := tmRe.FindAllStringSubmatch(prefix, -1)
			if len(tmMatches) > 0 {
				lastTm := tmMatches[len(tmMatches)-1]
				currentX, _ = strconv.ParseFloat(lastTm[5], 64)
				currentY, _ = strconv.ParseFloat(lastTm[6], 64)
			}
			tdMatches := tdRe.FindAllStringSubmatch(prefix, -1)
			if len(tdMatches) > 0 {
				lastTd := tdMatches[len(tdMatches)-1]
				dx, _ := strconv.ParseFloat(lastTd[1], 64)
				dy, _ := strconv.ParseFloat(lastTd[2], 64)
				currentX += dx
				currentY += dy
			}
			tfMatches := tfRe.FindAllStringSubmatch(prefix, -1)
			if len(tfMatches) > 0 {
				if fs, err := strconv.ParseFloat(tfMatches[len(tfMatches)-1][1], 64); err == nil && fs > 0 {
					currentFontSize = fs
				}
			}

			textStr := strings.TrimSpace(extractTextFromOperator(op))
			if textStr == "" {
				continue
			}
			height := currentFontSize
			if height < 8 {
				height = 8
			}
			width := float64(len([]rune(textStr))) * currentFontSize * 0.52
			if width < currentFontSize {
				width = currentFontSize
			}
			positions = append(positions, TextPosition{
				Text:   textStr,
				X:      currentX,
				Y:      currentY - (0.25 * height),
				Width:  width,
				Height: height,
			})
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
