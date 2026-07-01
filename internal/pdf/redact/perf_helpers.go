package redact

import (
	"bytes"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// bytesIndex finds sub in b without string conversion (PERF-32).
func bytesIndex(b, sub []byte) int {
	return bytes.Index(b, sub)
}

// parseObjectKey parses "num gen" object map keys without fmt.Sscanf.
func parseObjectKey(key string) (id, gen int, ok bool) {
	i := 0
	for i < len(key) && key[i] >= '0' && key[i] <= '9' {
		id = id*10 + int(key[i]-'0')
		i++
	}
	if i == 0 || i >= len(key) || key[i] != ' ' {
		return 0, 0, false
	}
	i++
	for i < len(key) && key[i] >= '0' && key[i] <= '9' {
		gen = gen*10 + int(key[i]-'0')
		i++
	}
	return id, gen, i == len(key)
}

// parseLeadingInt parses a decimal int from a byte slice prefix.
func parseLeadingInt(b []byte) (int, bool) {
	if len(b) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, false
	}
	return n, true
}

// splitFields splits on Unicode whitespace without strings.Fields allocation (PERF-186).
func splitFields(s string) []string {
	var parts []string
	start := -1
	for i, r := range s {
		if unicode.IsSpace(r) {
			if start >= 0 {
				parts = append(parts, s[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		parts = append(parts, s[start:])
	}
	return parts
}

// splitTabFields splits a TSV line on tabs without strings.Split (PERF-47).
func splitTabFields(line string) []string {
	if line == "" {
		return nil
	}
	var cols []string
	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] == '\t' {
			cols = append(cols, line[start:i])
			start = i + 1
		}
	}
	cols = append(cols, line[start:])
	return cols
}

// trimSpaceASCII trims ASCII whitespace in-place concept without full strings.TrimSpace alloc.
func trimSpaceASCII(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// equalsFoldASCII performs ASCII case-insensitive equality without ToLower allocation.
func equalsFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// containsFoldASCII checks if s contains substr case-insensitively (ASCII).
func containsFoldASCII(s, substr string) bool {
	if substr == "" {
		return true
	}
	sLen, subLen := len(s), len(substr)
	if subLen > sLen {
		return false
	}
	for i := 0; i+subLen <= sLen; i++ {
		if equalsFoldASCII(s[i:i+subLen], substr) {
			return true
		}
	}
	return false
}

// appendObjMapKey formats "num 0" map keys without fmt.Sprintf (PERF-6/109).
func appendObjMapKey(buf []byte, objNum int) []byte {
	buf = strconv.AppendInt(buf, int64(objNum), 10)
	buf = append(buf, ' ', '0')
	return buf
}

// appendHexUpper appends uppercase hex encoding without fmt.Sprintf.
const hexDigits = "0123456789ABCDEF"

func appendHexUpper(dst []byte, src []byte) []byte {
	for _, b := range src {
		dst = append(dst, hexDigits[b>>4], hexDigits[b&0x0F])
	}
	return dst
}

// writeHex4Runes writes a 4-digit uppercase hex rune to a strings.Builder.
func writeHex4Runes(sb *strings.Builder, r uint16) {
	sb.WriteByte(hexDigits[r>>12&0xF])
	sb.WriteByte(hexDigits[r>>8&0xF])
	sb.WriteByte(hexDigits[r>>4&0xF])
	sb.WriteByte(hexDigits[r&0xF])
}

// runeEqualFold compares rune slices case-insensitively without ToLower.
func runeEqualFold(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ra, rb := a[i], b[i]
		if ra == rb {
			continue
		}
		if ra >= 'A' && ra <= 'Z' {
			ra += 'a' - 'A'
		}
		if rb >= 'A' && rb <= 'Z' {
			rb += 'a' - 'A'
		}
		if ra != rb {
			return false
		}
	}
	return true
}

// normalizeSearchText joins trimmed lowercased fields without repeated ToLower in loops.
func normalizeSearchText(s string) string {
	s = trimSpaceASCII(s)
	if s == "" {
		return ""
	}
	fields := splitFields(strings.ToLower(s))
	if len(fields) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for i, f := range fields {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(f)
	}
	return b.String()
}

// appendPDFRect appends a PDF rectangle command without fmt.Sprintf.
func appendPDFRect(sb *strings.Builder, x, y, w, h float64) {
	appendFloat2(sb, x)
	sb.WriteByte(' ')
	appendFloat2(sb, y)
	sb.WriteByte(' ')
	appendFloat2(sb, w)
	sb.WriteByte(' ')
	appendFloat2(sb, h)
	sb.WriteString(" re f ")
}

func appendFloat2(sb *strings.Builder, f float64) {
	var buf [32]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', 2, 64)
	sb.Write(buf[:len(b)])
}

func appendFloat6(sb *strings.Builder, f float64) {
	var buf [32]byte
	b := strconv.AppendFloat(buf[:0], f, 'f', 6, 64)
	sb.Write(buf[:len(b)])
}

// utf16BEHexEscape writes PDF UTF-16BE hex escape sequences without fmt.Sprintf.
func utf16BEHexEscape(sb *strings.Builder, r rune) {
	writeEscaped := func(b byte) {
		sb.WriteString(`\x`)
		sb.WriteByte(hexDigits[b>>4])
		sb.WriteByte(hexDigits[b&0x0F])
	}
	if r <= 0xFFFF {
		writeEscaped(byte((r >> 8) & 0xFF))
		writeEscaped(byte(r & 0xFF))
		return
	}
	r -= 0x10000
	high := 0xD800 + ((r >> 10) & 0x3FF)
	low := 0xDC00 + (r & 0x3FF)
	writeEscaped(byte((high >> 8) & 0xFF))
	writeEscaped(byte(high & 0xFF))
	writeEscaped(byte((low >> 8) & 0xFF))
	writeEscaped(byte(low & 0xFF))
}

// stringToBytesForEncrypt copies a string into a reusable buffer for encryption APIs.
func stringToBytesForEncrypt(dst []byte, s string) []byte {
	n := len(s)
	if cap(dst) < n {
		dst = make([]byte, n)
	} else {
		dst = dst[:n]
	}
	copy(dst, s)
	return dst
}

// bytesEqualPrefix compares two byte slices with length precheck (PERF-48).
func bytesEqualPrefix(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a, b)
}

// joinLineParts builds "prev part" without += in loop (PERF-2).
func joinLineParts(prev, part string) string {
	if prev == "" {
		return part
	}
	var b strings.Builder
	b.Grow(len(prev) + len(part) + 1)
	b.WriteString(prev)
	b.WriteByte(' ')
	b.WriteString(part)
	return b.String()
}

// countUTF8Runes returns rune count without converting to []rune when possible.
func countUTF8Runes(s string) int {
	return utf8.RuneCountInString(s)
}

// appendXrefEntry formats a PDF xref subsection entry line.
func appendXrefEntry(offset, gen int) string {
	var b strings.Builder
	b.Grow(20)
	var scratch [16]byte
	off := strconv.AppendInt(scratch[:0], int64(offset), 10)
	for n := 10 - len(off); n > 0; n-- {
		b.WriteByte('0')
	}
	b.Write(off)
	b.WriteByte(' ')
	g := strconv.AppendInt(scratch[:0], int64(gen), 10)
	for n := 5 - len(g); n > 0; n-- {
		b.WriteByte('0')
	}
	b.Write(g)
	b.WriteString(" n \n")
	return b.String()
}

// appendPageWarning builds a page-scoped warning without strconv.Itoa allocations.
func appendPageWarning(pageNum int, msg string) string {
	var b strings.Builder
	b.Grow(16 + len(msg))
	b.WriteString("page ")
	var scratch [12]byte
	b.Write(strconv.AppendInt(scratch[:0], int64(pageNum), 10))
	b.WriteString(": ")
	b.WriteString(msg)
	return b.String()
}