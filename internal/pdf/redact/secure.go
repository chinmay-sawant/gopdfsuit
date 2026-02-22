package redact

import (
"bytes"
"compress/zlib"
"fmt"
"strings"
	"regexp"
	"errors"
	"math"

"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)
func (r *Redactor) applySecureContentRedactions(redactions []models.RedactionRect, queries []models.RedactionTextQuery) ([]byte, bool, []string, error) {
	objMap := r.objMap
	if objMap == nil {
		var err error
		objMap, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return nil, false, nil, err
		}
	}

	redactionsByPage := make(map[int][]models.RedactionRect)
	for _, r := range redactions {
		redactionsByPage[r.PageNum] = append(redactionsByPage[r.PageNum], r)
	}
	if len(redactionsByPage) == 0 && len(queries) > 0 {
		if info, err := r.GetPageInfo(); err == nil {
			for i := 1; i <= info.TotalPages; i++ {
				redactionsByPage[i] = nil
			}
		}
	}

	var warnings []string
	changedAny := false

	for pageNum, rects := range redactionsByPage {
		pageRef, err := findPageObject(objMap, r.pdfBytes, pageNum)
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

	out, err := rebuildPDF(objMap, r.pdfBytes)
	if err != nil {
		return nil, false, warnings, err
	}

	return out, changedAny, warnings, nil
}

func rewriteSecureStreamTree(objMap map[string][]byte, streamKey string, resources []byte, rects []models.RedactionRect, queries []models.RedactionTextQuery, visited map[string]bool) (bool, []string) {
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

func rewriteContentStreamSecure(streamObj []byte, rects []models.RedactionRect, queries []models.RedactionTextQuery) ([]byte, bool, error) {
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

func scrubDecodedContent(decoded []byte, rects []models.RedactionRect, queries []models.RedactionTextQuery) ([]byte, bool) {
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
				for _, r := range newText {
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

func applyRectMaskToText(text string, pos models.TextPosition, rects []models.RedactionRect) string {
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
