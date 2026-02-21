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
			redactions = append(redactions, findAllCombinedMatchRects(i, positions, normalizedQuery)...)
		}
	}
	return redactions, nil
}

// FindTextOccurrencesMulti searches for multiple terms and combines results.
func FindTextOccurrencesMulti(pdfBytes []byte, searchTexts []string) ([]RedactionRect, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	if len(searchTexts) == 0 {
		return nil, nil
	}

	seenTerms := make(map[string]struct{}, len(searchTexts))
	all := make([]RedactionRect, 0, len(searchTexts)*4)
	for _, raw := range searchTexts {
		term := strings.TrimSpace(raw)
		if term == "" {
			continue
		}
		key := strings.ToLower(term)
		if _, ok := seenTerms[key]; ok {
			continue
		}
		seenTerms[key] = struct{}{}

		rects, err := FindTextOccurrences(pdfBytes, term)
		if err != nil {
			return nil, err
		}
		all = append(all, rects...)
	}

	return all, nil
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

// findAllCombinedMatchRects finds ALL occurrences of normalizedQuery that span
// multiple text-show operators on the same visual line. It groups positions into
// lines (Y within half a character-height), concatenates each line's tokens in
// reading order, then scans for every non-overlapping match.
func findAllCombinedMatchRects(pageNum int, positions []TextPosition, normalizedQuery string) []RedactionRect {
	if len(positions) == 0 || normalizedQuery == "" {
		return nil
	}

	// Sort top-to-bottom then left-to-right (PDF Y is bottom-up so higher=first).
	ordered := append([]TextPosition(nil), positions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if math.Abs(ordered[i].Y-ordered[j].Y) < 3 {
			return ordered[i].X < ordered[j].X
		}
		return ordered[i].Y > ordered[j].Y
	})

	// Group into visual lines: tokens whose Y values are within half a glyph height
	// of the first token on that line belong together.
	type tokenSpan struct {
		pos   TextPosition
		start int
		end   int
	}
	type lineGroup struct {
		spans  []tokenSpan
		joined string
	}

	var lines []lineGroup
	for _, pos := range ordered {
		lineH := pos.Height
		if lineH <= 0 {
			lineH = 10
		}
		placed := false
		for li := range lines {
			if len(lines[li].spans) == 0 {
				continue
			}
			refY := lines[li].spans[0].pos.Y
			if math.Abs(pos.Y-refY) < lineH*0.75 {
				// Same line  — append token
				part := strings.TrimSpace(pos.Text)
				if part == "" {
					placed = true
					break
				}
				var startOff int
				if lines[li].joined == "" {
					startOff = 0
					lines[li].joined = part
				} else {
					startOff = len(lines[li].joined) + 1
					lines[li].joined += " " + part
				}
				lines[li].spans = append(lines[li].spans, tokenSpan{
					pos:   pos,
					start: startOff,
					end:   len(lines[li].joined),
				})
				placed = true
				break
			}
		}
		if !placed {
			part := strings.TrimSpace(pos.Text)
			if part == "" {
				lines = append(lines, lineGroup{})
				continue
			}
			lines = append(lines, lineGroup{
				spans:  []tokenSpan{{pos: pos, start: 0, end: len(part)}},
				joined: part,
			})
		}
	}

	var results []RedactionRect
	for _, line := range lines {
		if line.joined == "" || len(line.spans) < 2 {
			// Single-token lines are already handled by buildSubstringRects.
			continue
		}
		normalJoined := normalizeSearchText(line.joined)
		searchOff := 0
		for searchOff < len(normalJoined) {
			idx := strings.Index(normalJoined[searchOff:], normalizedQuery)
			if idx < 0 {
				break
			}
			matchStart := searchOff + idx
			matchEnd := matchStart + len(normalizedQuery)

			// Compute tight bounding box from only the overlapping tokens.
			minX := math.MaxFloat64
			minY := math.MaxFloat64
			maxX := -math.MaxFloat64
			maxY := -math.MaxFloat64
			for _, s := range line.spans {
				if s.start >= matchEnd || s.end <= matchStart {
					continue
				}
				// Partially-overlapping tokens: trim X proportionally using charW.
				charW := 0.0
				if s.end > s.start {
					charW = s.pos.Width / float64(s.end-s.start)
				}
				tokenX := s.pos.X
				tokenW := s.pos.Width
				if charW > 0 {
					overlapStart := matchStart - s.start
					if overlapStart < 0 {
						overlapStart = 0
					}
					overlapEnd := matchEnd - s.start
					if overlapEnd > s.end-s.start {
						overlapEnd = s.end - s.start
					}
					tokenX = s.pos.X + float64(overlapStart)*charW
					tokenW = float64(overlapEnd-overlapStart) * charW
				}
				if tokenX < minX {
					minX = tokenX
				}
				if s.pos.Y < minY {
					minY = s.pos.Y
				}
				if x := tokenX + tokenW; x > maxX {
					maxX = x
				}
				if y := s.pos.Y + s.pos.Height; y > maxY {
					maxY = y
				}
			}
			if minX < math.MaxFloat64 {
				results = append(results, RedactionRect{
					PageNum: pageNum,
					X:       minX,
					Y:       minY,
					Width:   maxX - minX,
					Height:  maxY - minY,
				})
			}
			// Advance past this match (non-overlapping).
			searchOff = matchEnd
		}
	}
	return results
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
	activeTextQueries := opts.TextSearch

	for _, q := range activeTextQueries {
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
		ocrRects, err := runOCRSearch(pdfBytes, activeTextQueries, *opts.OCR)
		if err != nil {
			report.Warnings = append(report.Warnings, "OCR fallback error: "+err.Error())
		} else {
			all = append(all, ocrRects...)
			report.MatchedTextCount += len(ocrRects)
		}
	}
	report.GeneratedRects = len(all)
	workingPDF := pdfBytes

	// Keep visual mode strictly visual to avoid mutating encoded text streams.
	// Best-effort secure rewriting can produce glyph corruption for complex font encodings.
	if mode == "visual_allowed" {
		report.Warnings = append(report.Warnings, "visual_allowed mode skipped secure text rewrite to preserve original glyph encoding")
	}

	if mode == "secure_required" {
		secureOut, secureChanged, secureWarns, err := applySecureContentRedactions(workingPDF, all, activeTextQueries)
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
		pageResources := findPageResources(pageBody, objMap)
		if len(keys) == 0 {
			warnings = append(warnings, fmt.Sprintf("page %d: no content streams", pageNum))
			continue
		}

		visited := make(map[string]bool)
		activeQueries := queries

		for _, key := range keys {
			changed, nestedWarnings := rewriteSecureStreamTree(objMap, key, pageResources, rects, activeQueries, visited)
			if changed {
				changedAny = true
			}
			warnings = append(warnings, nestedWarnings...)
		}
	}

	out, err := rebuildPDF(objMap, pdfBytes)
	if err != nil {
		return nil, false, warnings, err
	}

	return out, changedAny, warnings, nil
}

func rewriteSecureStreamTree(objMap map[string][]byte, streamKey string, resources []byte, rects []RedactionRect, queries []RedactionTextQuery, visited map[string]bool) (bool, []string) {
	if visited[streamKey] {
		return false, nil
	}
	visited[streamKey] = true

	objBody, ok := objMap[streamKey]
	if !ok {
		return false, nil
	}

	_, decoded, ok := inspectStream(objBody)
	if !ok {
		return false, nil
	}

	updated, changed, err := rewriteContentStreamSecure(objBody, rects, queries)
	warnings := make([]string, 0, 2)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("stream %s: %v", streamKey, err))
	} else if changed {
		objMap[streamKey] = updated
	}

	if len(resources) == 0 || len(decoded) == 0 {
		return changed, warnings
	}

	childRefs := resolveUsedXObjectRefs(decoded, resources)
	for _, childKey := range childRefs {
		childBody, ok := objMap[childKey]
		if !ok {
			continue
		}
		// Only recurse into Form XObjects where text content commonly lives.
		if !regexp.MustCompile(`/Subtype\s*/Form(\b|\s|/)`).Match(childBody) {
			continue
		}
		childResources := extractResourcesBody(childBody, objMap)
		childChanged, childWarnings := rewriteSecureStreamTree(objMap, childKey, childResources, rects, queries, visited)
		if childChanged {
			changed = true
		}
		warnings = append(warnings, childWarnings...)
	}

	return changed, warnings
}

func inspectStream(streamObj []byte) ([]byte, []byte, bool) {
	start, end, ok := locateStreamSegment(streamObj)
	if !ok {
		return nil, nil, false
	}
	raw := streamObj[start:end]
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
	start, end, ok := locateStreamSegment(streamObj)
	if !ok {
		return streamObj, false, nil
	}
	raw := streamObj[start:end]

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
	newObj = append(newObj, streamObj[:start]...)
	newObj = append(newObj, encoded...)
	newObj = append(newObj, streamObj[end:]...)

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
			trimmedOp := strings.TrimSpace(op)
			// Preserve the original encoding format.
			// CIDFont/Identity-H operators use <hex> Tj — re-encode replacement
			// text as UTF-16BE hex so the 2-byte code-pair structure stays intact.
			if strings.HasPrefix(trimmedOp, "<") && !strings.HasPrefix(trimmedOp, "[") {
				out.WriteString("<")
				for _, r := range []rune(newText) {
					_, _ = fmt.Fprintf(&out, "%04X", uint16(r))
				}
				out.WriteString("> Tj")
			} else {
				out.WriteString("(")
				out.WriteString(escapePDFTextLiteral(newText))
				out.WriteString(") Tj")
			}
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

	// In secure mode, if a redaction block covers a substantial portion of a text run,
	// scrub the full run. For small overlaps, keep per-glyph masking.
	for _, r := range rects {
		if !rectsIntersectWithTolerance(pos.X, pos.Y, pos.Width, pos.Height, r.X, r.Y, r.Width, r.Height, pos.Height*0.75) {
			continue
		}
		overlap := overlapWidth(pos.X, pos.Width, r.X, r.Width)
		coverage := overlap / pos.Width
		// Only blank the entire run when the rect covers ≥90% of it;
		// lower overlaps use per-glyph masking below to avoid over-redacting.
		if coverage >= 0.90 {
			for i := range runes {
				runes[i] = ' '
			}
			return string(runes)
		}
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

func rectsIntersectWithTolerance(x1, y1, w1, h1, x2, y2, w2, h2, pad float64) bool {
	if pad < 0 {
		pad = 0
	}
	return rectsIntersect(x1, y1, w1, h1, x2-pad, y2-pad, w2+(2*pad), h2+(2*pad))
}

func overlapWidth(x1, w1, x2, w2 float64) float64 {
	left := math.Max(x1, x2)
	right := math.Min(x1+w1, x2+w2)
	if right <= left {
		return 0
	}
	return right - left
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

func parseTextOperators(content []byte) []TextPosition {
	var positions []TextPosition

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
