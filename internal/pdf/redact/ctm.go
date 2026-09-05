package redact

import (
	"math"
	"strconv"
	"strings"
)

// Graphics-state (CTM) tracking for content streams.
//
// Real-world PDFs (notably Chromium-printed files) open their content with a
// flipped/scaled CTM such as `.24 0 0 -.24 0 792 cm` and draw everything
// top-down. Two redaction paths must account for the leftover state:
//
//   - The appended visual overlay inherits whatever CTM previous streams
//     left behind. Painting raw MediaBox numbers under a flipped CTM lands
//     the black box in the wrong place, so the overlay prefixes the inverse
//     matrix (netCTMInverse) to restore identity space first.
//   - Text-position extraction must compose the CTM at each text run;
//     otherwise secure masking intersects glyphs in the wrong space and
//     blanks unrelated spans.
//
// Canonical files (identity net CTM) behave exactly as before.

// affine is a 2D matrix [a b c d e f] mapping (x,y) to
// (a*x + c*y + e, b*x + d*y + f).
type affine [6]float64

var identityCTM = affine{1, 0, 0, 1, 0, 0}

// contentToken is one lexical unit of a content stream. Strings, hex
// strings, arrays, dicts, names, and inline-image blocks are returned as
// single opaque units so operator scanning never sees inside them.
type contentToken struct {
	kind       string // "word", "number", "string", "hex", "arrayOpen", "arrayClose", "dictOpen", "dictClose", "name", "other"
	text       string
	start, end int // byte offsets in the source for ordering against regex matches
}

// tokenizeContent splits raw (decoded) content-stream bytes into ordered
// tokens, skipping comments and inline-image data.
func tokenizeContent(content []byte) []contentToken {
	src := string(content)
	var tokens []contentToken
	i := 0
	emit := func(kind, text string, start, end int) {
		tokens = append(tokens, contentToken{kind: kind, text: text, start: start, end: end})
	}
	for i < len(src) {
		ch := src[i]
		switch {
		case ch == '%':
			for i < len(src) && src[i] != '\n' && src[i] != '\r' {
				i++
			}
		case ch == '(':
			start := i
			depth := 0
			for i < len(src) {
				if src[i] == '\\' {
					i += 2
					continue
				}
				if src[i] == '(' {
					depth++
				} else if src[i] == ')' {
					depth--
					i++
					if depth == 0 {
						break
					}
					continue
				}
				i++
			}
			emit("string", "", start, i)
		case ch == '<':
			start := i
			if i+1 < len(src) && src[i+1] == '<' {
				emit("dictOpen", "<<", start, start+2)
				i += 2
			} else {
				j := i + 1
				for j < len(src) && src[j] != '>' {
					j++
				}
				if j < len(src) {
					j++
				}
				i = j
				emit("hex", "", start, i)
			}
		case ch == '>':
			if i+1 < len(src) && src[i+1] == '>' {
				emit("dictClose", ">>", i, i+2)
				i += 2
			} else {
				i++
			}
		case ch == '[':
			emit("arrayOpen", "[", i, i+1)
			i++
		case ch == ']':
			emit("arrayClose", "]", i, i+1)
			i++
		case ch == '/':
			start := i
			j := i + 1
			for j < len(src) && !isContentDelim(src[j]) {
				j++
			}
			emit("name", src[start:j], start, j)
			i = j
		case isContentSpace(ch):
			i++
		default:
			start := i
			j := i
			for j < len(src) && !isContentDelim(src[j]) {
				j++
			}
			word := src[i:j]
			if word == "BI" {
				// Inline image: skip everything through the EI terminator.
				k := j
				for k < len(src) {
					if src[k] == 'E' && k+1 < len(src) && src[k+1] == 'I' &&
						(k+2 >= len(src) || isContentSpace(src[k+2])) &&
						(k == 0 || isContentSpace(src[k-1])) {
						k += 2
						break
					}
					k++
				}
				i = k
				continue
			}
			i = j
			if _, err := strconv.ParseFloat(word, 64); err == nil {
				emit("number", word, start, i)
			} else {
				emit("word", word, start, i)
			}
		}
	}
	return tokens
}

func isContentSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f' || ch == 0
}

func isContentDelim(ch byte) bool {
	return isContentSpace(ch) || ch == '(' || ch == ')' || ch == '<' || ch == '>' || ch == '[' || ch == ']' || ch == '/' || ch == '%'
}

func composeCTM(base, next affine) affine {
	a1, b1, c1, d1, e1, f1 := base[0], base[1], base[2], base[3], base[4], base[5]
	a2, b2, c2, d2, e2, f2 := next[0], next[1], next[2], next[3], next[4], next[5]
	return affine{
		a1*a2 + c1*b2,
		b1*a2 + d1*b2,
		a1*c2 + c1*d2,
		b1*c2 + d1*d2,
		a1*e2 + c1*f2 + e1,
		b1*e2 + d1*f2 + f1,
	}
}

// netCTM folds q/Q/cm across decoded content streams in order, yielding the
// graphics state the next appended stream would inherit.
func netCTM(streams [][]byte) affine {
	current := identityCTM
	var stack []affine
	for _, stream := range streams {
		for _, ev := range graphicsEvents(stream) {
			current, stack = applyGraphicsEvent(current, stack, ev)
		}
	}
	return current
}

func applyGraphicsEvent(current affine, stack []affine, ev graphicsEvent) (affine, []affine) {
	switch ev.op {
	case "q":
		return current, append(stack, current)
	case "Q":
		if len(stack) > 0 {
			current = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		}
		return current, stack
	case "cm":
		return composeCTM(current, affine(ev.nums)), stack
	default:
		return current, stack
	}
}

// graphicsEvent is a CTM-affecting operator with its source offset.
type graphicsEvent struct {
	end  int
	op   string // "q", "Q", "cm"
	nums [6]float64
}

// graphicsEvents extracts q/Q/cm operators in source order. Numbers
// immediately preceding cm are captured; anything else resets the operand
// run so a cm can never grab unrelated numbers.
func graphicsEvents(content []byte) []graphicsEvent {
	var events []graphicsEvent
	var numbers []float64
	for _, token := range tokenizeContent(content) {
		switch token.kind {
		case "number":
			if value, err := strconv.ParseFloat(token.text, 64); err == nil {
				numbers = append(numbers, value)
			}
		case "word":
			switch token.text {
			case "q", "Q":
				events = append(events, graphicsEvent{end: token.end, op: token.text})
				numbers = numbers[:0]
			case "cm":
				if len(numbers) >= 6 {
					var nums [6]float64
					copy(nums[:], numbers[len(numbers)-6:])
					events = append(events, graphicsEvent{end: token.end, op: "cm", nums: nums})
				}
				numbers = numbers[:0]
			default:
				numbers = numbers[:0]
			}
		default:
			numbers = numbers[:0]
		}
	}
	return events
}

// ctmAtOffset replays graphics events ending at or before offset,
// yielding the CTM in force at that source position.
func ctmAtOffset(events []graphicsEvent, offset int) affine {
	current := identityCTM
	var stack []affine
	for _, ev := range events {
		if ev.end > offset {
			break
		}
		current, stack = applyGraphicsEvent(current, stack, ev)
	}
	return current
}

func invertAffine(m affine) (affine, bool) {
	det := m[0]*m[3] - m[1]*m[2]
	if math.Abs(det) < 1e-12 {
		return identityCTM, false
	}
	return affine{
		m[3] / det,
		-m[1] / det,
		-m[2] / det,
		m[0] / det,
		(m[2]*m[5] - m[3]*m[4]) / det,
		(m[1]*m[4] - m[0]*m[5]) / det,
	}, true
}

func isIdentityCTM(m affine) bool {
	const eps = 1e-9
	want := identityCTM
	for i := range m {
		if math.Abs(m[i]-want[i]) > eps {
			return false
		}
	}
	return true
}

func applyAffine(m affine, x, y float64) (float64, float64) {
	return m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]
}

// formatAffine renders a matrix for emission into a content stream.
func formatAffine(m affine) string {
	var sb strings.Builder
	for i, value := range m {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(strconv.FormatFloat(value, 'f', 4, 64))
	}
	return sb.String()
}

// decodedPageStreams returns the decoded bytes of a page's content streams
// in order, for graphics-state analysis.
func decodedPageStreams(objMap map[int][]byte, pageBody []byte) [][]byte {
	var out [][]byte
	for _, key := range extractContentKeys(pageBody) {
		body, ok := objMap[key]
		if !ok {
			continue
		}
		_, decoded, ok := inspectStream(body)
		if !ok || len(decoded) == 0 {
			continue
		}
		out = append(out, decoded)
	}
	return out
}

// netCTMInverse returns the printable inverse of the graphics state a
// page's existing streams leave behind, or false when it is already the
// identity. An appended overlay executes under that leftover state, so it
// must prefix the inverse to paint in default user space.
func netCTMInverse(objMap map[int][]byte, pageBody []byte) (string, bool) {
	net := netCTM(decodedPageStreams(objMap, pageBody))
	if isIdentityCTM(net) {
		return "", false
	}
	inv, ok := invertAffine(net)
	if !ok {
		return "", false
	}
	return formatAffine(inv), true
}
