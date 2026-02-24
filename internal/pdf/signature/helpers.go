package signature

import (
	"fmt"
	"strings"
)

// fmtNum formats a float64 to a compact string, stripping trailing zeros.
func fmtNum(f float64) string {
	s := fmt.Sprintf("%.4f", f)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// escapeText escapes parentheses and backslashes for PDF text strings.
func escapeText(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	return s
}
