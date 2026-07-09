package redact

import (
	"bytes"
	"compress/zlib"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/byteconv"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

func isAllWS(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// Package-level regexes (PERF-1): avoid MustCompile on hot paths / in loops.
var (
	secureFormSubtypeRe = regexp.MustCompile(`/Subtype\s*/Form(\b|\s|/)`)
	secureContentsRe    = regexp.MustCompile(`/Contents\s+(?:(\d+)\s+(\d+)\s+R|\[(.*?)\])`)
	secureRefRe         = regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	secureLenRe         = regexp.MustCompile(`/Length\s+(?:\d+\s+\d+\s+R|\d+)`)
	secureTextOpRe      = regexp.MustCompile(`(?s)\[(?:.|\n|\r)*?\]\s*TJ|<[^>]+>\s*Tj|\((?:\\.|[^\\)])*\)\s*Tj|\((?:\\.|[^\\)])*\)\s*'|[\d.-]+\s+[\d.-]+\s+\((?:\\.|[^\\)])*\)\s*"`)
)

// PERF-227/233: pool of zlib writers with BestSpeed compression for the hot path.
var zlibWriterPool = sync.Pool{
	New: func() any {
		w, _ := zlib.NewWriterLevel(nil, zlib.BestSpeed)
		return w
	},
}

func (r *Redactor) applySecureContentRedactions(redactions []models.RedactionRect, queries []models.RedactionTextQuery) ([]byte, bool, []string, error) {
	objMap := r.objMap
	objGen := r.objGen
	if objMap == nil {
		var err error
		objMap, objGen, err = buildObjectMap(r.pdfBytes)
		if err != nil {
			return nil, false, nil, err
		}
	}

	redactionsByPage := make(map[int][]models.RedactionRect, len(redactions)) // PERF-192
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
		pageObjNum, err := findPageObject(objMap, r.pdfBytes, pageNum)
		if err != nil {
			// PERF-6: avoid fmt.Sprintf in loop
			var wb strings.Builder
			wb.Grow(64)
			var tmp [20]byte
			wb.WriteString("page ")
			wb.Write(strconv.AppendInt(tmp[:0], int64(pageNum), 10))
			wb.WriteString(": ")
			wb.WriteString(err.Error())
			warnings = append(warnings, wb.String())
			continue
		}
		pageBody := objMap[pageObjNum]
		keys := extractContentKeys(pageBody)
		pageResources := findPageResources(pageBody, objMap)
		if len(keys) == 0 {
			var wb strings.Builder
			wb.Grow(64)
			var tmp [20]byte
			wb.WriteString("page ")
			wb.Write(strconv.AppendInt(tmp[:0], int64(pageNum), 10))
			wb.WriteString(": no content streams")
			warnings = append(warnings, wb.String())
			continue
		}

		visited := make(map[int]bool, 8) // PERF-192
		activeQueries := queries

		for _, key := range keys {
			changed, nestedWarnings := rewriteSecureStreamTree(objMap, key, pageResources, rects, activeQueries, visited)
			if changed {
				changedAny = true
			}
			warnings = append(warnings, nestedWarnings...)
		}
	}

	out, err := rebuildPDF(objMap, objGen, r.pdfBytes)
	if err != nil {
		return nil, false, warnings, err
	}

	return out, changedAny, warnings, nil
}

func rewriteSecureStreamTree(objMap map[int][]byte, streamObjNum int, resources []byte, rects []models.RedactionRect, queries []models.RedactionTextQuery, visited map[int]bool) (bool, []string) {
	if visited[streamObjNum] {
		return false, nil
	}
	visited[streamObjNum] = true

	objBody, ok := objMap[streamObjNum]
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
		var wb strings.Builder
		wb.Grow(64)
		var tmp [20]byte
		wb.WriteString("stream ")
		wb.Write(strconv.AppendInt(tmp[:0], int64(streamObjNum), 10))
		wb.WriteString(": ")
		wb.WriteString(err.Error())
		warnings = append(warnings, wb.String())
	} else if changed {
		objMap[streamObjNum] = updated
	}

	if len(resources) == 0 || len(decoded) == 0 {
		return changed, warnings
	}

	childRefs := resolveUsedXObjectRefs(decoded, resources)
	for _, childNum := range childRefs {
		childBody, ok := objMap[childNum]
		if !ok {
			continue
		}
		// Only recurse into Form XObjects where text content commonly lives.
		if !secureFormSubtypeRe.Match(childBody) {
			continue
		}
		childResources := extractResourcesBody(childBody, objMap)
		childChanged, childWarnings := rewriteSecureStreamTree(objMap, childNum, childResources, rects, queries, visited)
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

func extractContentKeys(pageBody []byte) []int {
	match := secureContentsRe.FindSubmatch(pageBody)
	if match == nil {
		return nil
	}
	var keys []int
	if len(match[1]) > 0 {
		n, err := strconv.Atoi(string(match[1]))
		if err == nil {
			keys = append(keys, n)
		}
		return keys
	}
	if len(match[3]) > 0 {
		refs := secureRefRe.FindAllSubmatch(match[3], -1)
		for _, r := range refs {
			n, err := strconv.Atoi(string(r[1]))
			if err == nil {
				keys = append(keys, n)
			}
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
		zw := zlibWriterPool.Get().(*zlib.Writer)
		zw.Reset(&buf)
		if _, err := zw.Write(newDecoded); err != nil {
			zw.Close()
			zlibWriterPool.Put(zw)
			return nil, false, err
		}
		if err := zw.Close(); err != nil {
			zlibWriterPool.Put(zw)
			return nil, false, err
		}
		zlibWriterPool.Put(zw)
		encoded = buf.Bytes()
	}

	// PERF-119: single multi-append of stream segments
	newObj := make([]byte, 0, len(streamObj)-(end-start)+len(encoded)+64)
	newObj = append(append(append(newObj, streamObj[:start]...), encoded...), streamObj[end:]...)

	// PERF-32/6: build /Length without fmt
	var lenBuf [32]byte
	n := copy(lenBuf[:], "/Length ")
	bLen := strconv.AppendInt(lenBuf[n:n], int64(len(encoded)), 10)
	n += len(bLen)
	newObj = secureLenRe.ReplaceAll(newObj, lenBuf[:n])

	return newObj, true, nil
}

func scrubDecodedContent(decoded []byte, rects []models.RedactionRect, queries []models.RedactionTextQuery) ([]byte, bool) {
	positions := parseTextOperators(decoded)

	src := string(decoded)
	matches := secureTextOpRe.FindAllStringIndex(src, -1)
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
		extracted := extractTextFromOperator(op)
		if isAllWS(extracted) {
			out.WriteString(op)
			last = m[1]
			continue
		}
		text := trimSpace(extracted)

		newText := text
		if posIdx < len(positions) {
			p := positions[posIdx]
			// Compare after trim only when needed (PERF-46)
			if pText := trimSpace(p.Text); pText != text {
				for lookahead := posIdx + 1; lookahead < len(positions) && lookahead < posIdx+6; lookahead++ {
					if trimSpace(positions[lookahead].Text) == text {
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
			if isAllWS(q.Text) {
				continue
			}
			term := trimSpace(q.Text)
			newText = replaceCaseInsensitiveWithSpaces(newText, term)
		}

		if newText != text {
			changed = true
			trimmedOp := trimSpace(op)
			// CIDFont/Identity-H operators use <hex> Tj encoding.
			isHex := strings.HasPrefix(trimmedOp, "<") && !strings.HasPrefix(trimmedOp, "[")
			// Use a TJ array with kerning adjustments so that remaining
			// text stays in its original position after characters are
			// removed. Simple Tj with spaces causes text to shift because
			// space glyphs are narrower than letter glyphs.
			out.WriteString(buildRedactionTJArray(text, newText, isHex))
		} else {
			out.WriteString(op)
		}
		last = m[1]
	}

	out.WriteString(src[last:])
	if !changed {
		return decoded, false
	}
	return byteconv.StringToBytes(out.String()), true
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

// buildRedactionTJArray constructs a TJ array operator that uses kerning
// adjustments so that text remaining after redaction stays in position.
// Each removed character is replaced by a kern of -520 units (matching
// the 0.52*fontSize heuristic used in parseTextOperators). This avoids
// the problem of space glyphs being narrower than letter glyphs.
// appendHex4 appends a 4-digit uppercase hex value without fmt (PERF-6).
func appendHex4(b *strings.Builder, v uint16) {
	const hexdigits = "0123456789ABCDEF"
	b.WriteByte(hexdigits[v>>12])
	b.WriteByte(hexdigits[(v>>8)&0xf])
	b.WriteByte(hexdigits[(v>>4)&0xf])
	b.WriteByte(hexdigits[v&0xf])
}

func buildRedactionTJArray(original, redacted string, isHex bool) string {
	origRunes := []rune(original)
	redRunes := []rune(redacted)

	// If lengths differ unexpectedly, fall back to simple Tj.
	if len(origRunes) != len(redRunes) {
		if isHex {
			var sb strings.Builder
			sb.Grow(2 + len(redacted)*4 + 4)
			sb.WriteString("<")
			for _, r := range redacted {
				appendHex4(&sb, uint16(r))
			}
			sb.WriteString("> Tj")
			return sb.String()
		}
		return "(" + escapePDFTextLiteral(redacted) + ") Tj"
	}

	// Check if any character was actually replaced (non-space → space).
	hasRedaction := false
	for i := range origRunes {
		if origRunes[i] != ' ' && redRunes[i] == ' ' {
			hasRedaction = true
			break
		}
	}
	if !hasRedaction {
		// No positional redaction: use simple Tj format.
		if isHex {
			var sb strings.Builder
			sb.Grow(2 + len(redacted)*4 + 4)
			sb.WriteString("<")
			for _, r := range redacted {
				appendHex4(&sb, uint16(r))
			}
			sb.WriteString("> Tj")
			return sb.String()
		}
		return "(" + escapePDFTextLiteral(redacted) + ") Tj"
	}

	// Build segments: text runs and kern (removed-char) runs.
	type segment struct {
		text    string
		removed string // chars that were removed (for kerning estimation)
		isKern  bool
	}

	var segments []segment
	var textBuf, kernBuf strings.Builder

	for i := range redRunes {
		wasReplaced := origRunes[i] != ' ' && redRunes[i] == ' '
		if wasReplaced {
			if textBuf.Len() > 0 {
				segments = append(segments, segment{text: textBuf.String()})
				textBuf.Reset()
			}
			kernBuf.WriteRune(origRunes[i])
		} else {
			if kernBuf.Len() > 0 {
				segments = append(segments, segment{isKern: true, removed: kernBuf.String()})
				kernBuf.Reset()
			}
			textBuf.WriteRune(redRunes[i])
		}
	}
	if kernBuf.Len() > 0 {
		segments = append(segments, segment{isKern: true, removed: kernBuf.String()})
	}
	if textBuf.Len() > 0 {
		segments = append(segments, segment{text: textBuf.String()})
	}

	// Build TJ array: [(text) kern (text) ...] TJ
	var out strings.Builder
	out.Grow(len(segments) * 16)
	out.WriteString("[")
	for _, seg := range segments {
		if seg.isKern {
			// Estimate width of removed string
			// TJ array kerning units are 1/1000 of text space.
			estWidth := estimateStringWidth(seg.removed, 1000)
			// Negative value = advance cursor to the right.
			kern := -int(math.Round(estWidth))
			var kernTmp [20]byte
			out.Write(strconv.AppendInt(kernTmp[:0], int64(kern), 10))
			out.WriteByte(' ')
		} else {
			if isHex {
				out.WriteString("<")
				for _, r := range seg.text {
					appendHex4(&out, uint16(r))
				}
				out.WriteString("> ")
			} else {
				out.WriteString("(")
				out.WriteString(escapePDFTextLiteral(seg.text))
				out.WriteString(") ")
			}
		}
	}
	out.WriteString("] TJ")
	return out.String()
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
