package redact

import (
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// ExtractTextPositions extracts text with coordinates from a specific page
func (r *Redactor) ExtractTextPositions(pageNum int) ([]models.TextPosition, error) {
	if len(r.pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}

	objMap := r.objMap
	if objMap == nil {
		var err error
		objMap, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return nil, err
		}
	}

	pageRef, err := findPageObject(objMap, r.pdfBytes, pageNum)
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
func (r *Redactor) FindTextOccurrences(searchText string) ([]models.RedactionRect, error) {
	if len(r.pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	if searchText == "" {
		return nil, nil
	}

	info, err := r.GetPageInfo()
	if err != nil {
		return nil, err
	}

	var redactions []models.RedactionRect
	normalizedQuery := r.normalizeSearchText(searchText)
	searchText = strings.ToLower(searchText)

	for i := 1; i <= info.TotalPages; i++ {
		positions, err := r.ExtractTextPositions(i)
		if err != nil {
			// Log error but continue? Or fail? Best to continue for search flexibility
			continue
		}

		for _, pos := range positions {
			redactions = append(redactions, r.buildSubstringRects(i, pos, searchText)...)
		}

		// Fallback for PDFs that split words/phrases across multiple text-show operators
		// (e.g., "don" + "ald" as two separate Tj ops, or "Jeffrey" + "Epstein").
		// The guard inside r.findAllCombinedMatchRects skips single-token lines already
		// handled by r.buildSubstringRects above.
		if len(positions) > 1 {
			redactions = append(redactions, r.findAllCombinedMatchRects(i, positions, normalizedQuery)...)
		}
	}
	return redactions, nil
}

// FindTextOccurrencesMulti searches for multiple terms and combines results.
func (r *Redactor) FindTextOccurrencesMulti(searchTexts []string) ([]models.RedactionRect, error) {
	if len(r.pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}
	if len(searchTexts) == 0 {
		return nil, nil
	}

	seenTerms := make(map[string]struct{}, len(searchTexts))
	all := make([]models.RedactionRect, 0, len(searchTexts)*4)
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

		rects, err := r.FindTextOccurrences(term)
		if err != nil {
			return nil, err
		}
		all = append(all, rects...)
	}

	return all, nil
}

// isURLToken returns true when a text token is a URL or URL fragment.
// Proportional character-offset estimation is unreliable for these because
// they are packed with narrow chars (:, /, ., -, &, ?, =) that skew the
// average glyph width. In secure_required mode the content stream rewrite
// already scrubs the text; a wrong-position overlay just confuses the output.
func (r *Redactor) isURLToken(text string) bool {
	if strings.Contains(text, "://") {
		return true
	}
	runes := []rune(text)
	if len(runes) <= 30 || strings.ContainsRune(text, ' ') {
		return false
	}
	// URL query / path fragments: no spaces, longer than 30 chars, contain
	// multiple URL-special characters (&, =, +, %).
	queryCount := 0
	hyphenCount := 0
	for _, ch := range text {
		switch ch {
		case '&', '=', '+', '%', '?':
			queryCount++
		case '-':
			hyphenCount++
		}
	}
	if queryCount >= 2 {
		return true
	}
	// URL path slug: long no-space token with many hyphens (e.g. wrapped URL
	// path lines like "birther-wagon-insists-Hillary-drove-says-lot-problems-Bill-Clinton-s-").
	if len(runes) > 40 && hyphenCount >= 4 {
		return true
	}
	return false
}

func (r *Redactor) buildSubstringRects(pageNum int, pos models.TextPosition, loweredSearch string) []models.RedactionRect {
	if loweredSearch == "" || strings.TrimSpace(pos.Text) == "" {
		return nil
	}
	src := []rune(strings.ToLower(pos.Text))
	needle := []rune(loweredSearch)
	if len(src) == 0 || len(needle) == 0 || len(needle) > len(src) {
		return nil
	}

	// Original string preserved for precise width estimation (case-sensitive)
	origSrc := []rune(pos.Text)
	totalEst := estimateStringWidth(pos.Text, pos.Height)
	scale := 1.0
	if totalEst > 0 && pos.Width > 0 {
		scale = pos.Width / totalEst
	}

	urlToken := r.isURLToken(pos.Text)

	rects := make([]models.RedactionRect, 0, 2)
	for i := 0; i+len(needle) <= len(src); i++ {
		if !r.runeSliceEqual(src[i:i+len(needle)], needle) {
			continue
		}
		if urlToken {
			// Redact the entire URL token with one rect and stop scanning.
			return []models.RedactionRect{{
				PageNum: pageNum,
				X:       pos.X,
				Y:       pos.Y,
				Width:   pos.Width,
				Height:  pos.Height,
			}}
		}

		var offsetEst, matchEst float64
		for j := 0; j < i; j++ {
			if j < len(origSrc) {
				offsetEst += estimateStringWidth(string(origSrc[j]), pos.Height)
			}
		}
		for j := 0; j < len(needle); j++ {
			if i+j < len(origSrc) {
				matchEst += estimateStringWidth(string(origSrc[i+j]), pos.Height)
			}
		}

		x := pos.X + (offsetEst * scale)
		w := matchEst * scale

		rects = append(rects, models.RedactionRect{
			PageNum: pageNum,
			X:       x,
			Y:       pos.Y,
			Width:   w,
			Height:  pos.Height,
		})
	}
	return rects
}

func (r *Redactor) runeSliceEqual(a, b []rune) bool {
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

func (r *Redactor) normalizeSearchText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

// r.findAllCombinedMatchRects finds ALL occurrences of normalizedQuery that span
// multiple text-show operators on the same visual line. It groups positions into
// lines (Y within half a character-height), concatenates each line's tokens in
// reading order, then scans for every non-overlapping match.
//nolint:gocyclo
func (r *Redactor) findAllCombinedMatchRects(pageNum int, positions []models.TextPosition, normalizedQuery string) []models.RedactionRect {
	if len(positions) == 0 || normalizedQuery == "" {
		return nil
	}

	// Sort top-to-bottom then left-to-right (PDF Y is bottom-up so higher=first).
	ordered := append([]models.TextPosition(nil), positions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if math.Abs(ordered[i].Y-ordered[j].Y) < 3 {
			return ordered[i].X < ordered[j].X
		}
		return ordered[i].Y > ordered[j].Y
	})

	// Group into visual lines: tokens whose Y values are within half a glyph height
	// of the first token on that line belong together.
	type tokenSpan struct {
		pos   models.TextPosition
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

	var results []models.RedactionRect
	for _, line := range lines {
		if line.joined == "" || len(line.spans) < 2 {
			// Single-token lines are already handled by r.buildSubstringRects.
			continue
		}
		normalJoined := r.normalizeSearchText(line.joined)
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
				// URL token: redact the whole token — proportional offset is
				// unreliable for these, but the token itself must be covered.
				if r.isURLToken(s.pos.Text) {
					if s.pos.X < minX {
						minX = s.pos.X
					}
					if s.pos.Y < minY {
						minY = s.pos.Y
					}
					if x := s.pos.X + s.pos.Width; x > maxX {
						maxX = x
					}
					if y := s.pos.Y + s.pos.Height; y > maxY {
						maxY = y
					}
					continue
				}
				// Partially-overlapping tokens: trim X proportionally using estimated widths.
				tokenX := s.pos.X
				tokenW := s.pos.Width
				tokenText := []rune(s.pos.Text)
				tokenEst := estimateStringWidth(s.pos.Text, s.pos.Height)

				if tokenEst > 0 && s.pos.Width > 0 && len(tokenText) > 0 {
					scale := s.pos.Width / tokenEst

					overlapStart := matchStart - s.start
					if overlapStart < 0 {
						overlapStart = 0
					}

					overlapEnd := matchEnd - s.start
					if overlapEnd > s.end-s.start {
						overlapEnd = s.end - s.start
					}

					var offsetEst, matchEst float64
					for j := 0; j < overlapStart; j++ {
						if j < len(tokenText) {
							offsetEst += estimateStringWidth(string(tokenText[j]), s.pos.Height)
						}
					}
					for j := overlapStart; j < overlapEnd; j++ {
						if j < len(tokenText) {
							matchEst += estimateStringWidth(string(tokenText[j]), s.pos.Height)
						}
					}

					tokenX = s.pos.X + (offsetEst * scale)
					tokenW = matchEst * scale
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
				results = append(results, models.RedactionRect{
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
