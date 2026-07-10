package redact

import (
	"errors"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

// ExtractTextPositions extracts text with coordinates from a specific page
func (r *Redactor) ExtractTextPositions(pageNum int) ([]models.TextPosition, error) {
	if len(r.pdfBytes) == 0 {
		return nil, errors.New("empty pdf bytes")
	}

	objMap := r.objMap
	if objMap == nil {
		var err error
		objMap, _, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return nil, err
		}
	}

	pageObjNum, err := findPageObject(objMap, r.pdfBytes, pageNum)
	if err != nil {
		return nil, err
	}

	pageBody := objMap[pageObjNum]
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
	var lowered string
	for _, raw := range searchTexts {
		// Skip empty/whitespace without always allocating TrimSpace (PERF-46)
		if len(raw) == 0 {
			continue
		}
		term := trimSpace(raw)
		if term == "" {
			continue
		}
		lowered = strings.ToLower(term) // PERF-109: cache map key outside loop
		if _, ok := seenTerms[lowered]; ok {
			continue
		}
		seenTerms[lowered] = struct{}{}

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
	if loweredSearch == "" || trimSpace(pos.Text) == "" {
		return nil
	}
	src := []rune(pos.Text)
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

	// PERF-230: Precompute per-rune widths for the height outside the scan loop.
	origWidths := make([]float64, len(origSrc))
	for j := range origSrc {
		origWidths[j] = pos.Height * runeWidthFactor(origSrc[j])
	}

	urlToken := r.isURLToken(pos.Text)

	rects := make([]models.RedactionRect, 0, 2)
	for i := 0; i+len(needle) <= len(src); i++ {
		if !r.runeSliceEqualFold(src[i:i+len(needle)], needle) {
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
			offsetEst += origWidths[j]
		}
		for j := 0; j < len(needle); j++ {
			matchEst += origWidths[i+j]
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

func runeWidthFactor(r rune) float64 {
	switch r {
	case 'i', 'j', 'l', 'I', '1', '.', ',', ';', ':', '!', '\'', '|':
		return 0.25
	case 'f', 't', 'r', '-', ' ', '(', ')':
		return 0.35
	case 'm', 'w', 'M', 'W', 'O', 'Q', '@', '%':
		return 0.8
	default:
		switch {
		case r >= 'A' && r <= 'Z':
			return 0.65
		case r >= '0' && r <= '9':
			return 0.55
		default:
			return 0.52
		}
	}
}

func (r *Redactor) runeSliceEqualFold(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if unicode.SimpleFold(a[i]) != unicode.SimpleFold(b[i]) && a[i] != b[i] {
			ra, rb := unicode.ToLower(a[i]), unicode.ToLower(b[i])
			if ra != rb {
				return false
			}
		}
	}
	return true
}

// normalizeSearchText lowercases and collapses whitespace in a single pass (PERF-2).
func (r *Redactor) normalizeSearchText(s string) string {
	s = trimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, rr := range s {
		if unicode.IsSpace(rr) {
			if !prevSpace && b.Len() > 0 {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(unicode.ToLower(rr))
	}
	out := b.String()
	if len(out) > 0 && out[len(out)-1] == ' ' {
		out = out[:len(out)-1]
	}
	return out
}

// r.findAllCombinedMatchRects finds ALL occurrences of normalizedQuery that span
// multiple text-show operators on the same visual line. It groups positions into
// lines (Y within half a character-height), concatenates each line's tokens in
// reading order, then scans for every non-overlapping match.
//
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
	// Pointer groups so strings.Builder is never copied when the slice grows (PERF-2).
	type lineGroup struct {
		spans  []tokenSpan
		joined strings.Builder
	}

	var lines []lineGroup
	for _, pos := range ordered {
		lineH := pos.Height
		if lineH <= 0 {
			lineH = 10
		}
		placed := false
		for li := range lines {
			line := &lines[li]
			if len(line.spans) == 0 {
				continue
			}
			refY := line.spans[0].pos.Y
			if math.Abs(pos.Y-refY) < lineH*0.75 {
				// Same line  — append token
				part := trimSpace(pos.Text)
				if part == "" {
					placed = true
					break
				}
				var startOff int
				if line.joined.Len() == 0 {
					startOff = 0
					line.joined.WriteString(part)
				} else {
					startOff = line.joined.Len() + 1
					line.joined.WriteByte(' ')
					line.joined.WriteString(part)
				}
				line.spans = append(line.spans, tokenSpan{
					pos:   pos,
					start: startOff,
					end:   line.joined.Len(),
				})
				placed = true
				break
			}
		}
		if !placed {
			part := trimSpace(pos.Text)
			if part == "" {
				lines = append(lines, lineGroup{})
				continue
			}
			lg := lineGroup{
				spans: []tokenSpan{{pos: pos, start: 0, end: len(part)}},
			}
			lg.joined.WriteString(part)
			lines = append(lines, lg)
		}
	}

	var results []models.RedactionRect
	var tokenWidths []float64
	var spanTokenEst []float64
	for li := range lines {
		line := &lines[li]
		if line.joined.Len() == 0 || len(line.spans) < 2 {
			// Single-token lines are already handled by r.buildSubstringRects.
			continue
		}
		normalJoined := r.normalizeSearchText(line.joined.String())
		if cap(spanTokenEst) < len(line.spans) {
			spanTokenEst = make([]float64, len(line.spans))
		}
		spanTokenEst = spanTokenEst[:len(line.spans)]
		for i, s := range line.spans {
			spanTokenEst[i] = estimateStringWidth(s.pos.Text, s.pos.Height)
		}
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
			for si, s := range line.spans {
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
				tokenEst := spanTokenEst[si]

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

					// Precompute per-rune widths outside the inner scan loops.
					if cap(tokenWidths) < len(tokenText) {
						tokenWidths = make([]float64, len(tokenText))
					}
					tokenWidths = tokenWidths[:len(tokenText)]
					for j := range tokenText {
						tokenWidths[j] = s.pos.Height * runeWidthFactor(tokenText[j])
					}

					var offsetEst, matchEst float64
					for j := 0; j < overlapStart && j < len(tokenWidths); j++ {
						offsetEst += tokenWidths[j]
					}
					for j := overlapStart; j < overlapEnd && j < len(tokenWidths); j++ {
						matchEst += tokenWidths[j]
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
