package signature

import (
	"fmt"
	"strconv"
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

func writePDFDictText(sb *strings.Builder, key, text string) {
	sb.WriteString(key)
	sb.WriteString(" (")
	sb.WriteString(escapeText(text))
	sb.WriteByte(')')
}

func writePDFRect(sb *strings.Builder, x, y, w, h float64) {
	sb.WriteString(" /Rect [")
	sb.WriteString(fmtNum(x))
	sb.WriteByte(' ')
	sb.WriteString(fmtNum(y))
	sb.WriteByte(' ')
	sb.WriteString(fmtNum(x + w))
	sb.WriteByte(' ')
	sb.WriteString(fmtNum(y + h))
	sb.WriteByte(']')
}

func writeFormRectCmd(sb *strings.Builder, w, h float64, suffix string) {
	sb.WriteString("0 0 ")
	sb.WriteString(fmtNum(w))
	sb.WriteByte(' ')
	sb.WriteString(fmtNum(h))
	sb.WriteString(" re ")
	sb.WriteString(suffix)
}

func writeObjRef(sb *strings.Builder, objID int) {
	var buf [16]byte
	sb.Write(strconv.AppendInt(buf[:0], int64(objID), 10))
	sb.WriteString(" 0 R")
}

func writePadded2(sb *strings.Builder, v int) {
	if v < 10 {
		sb.WriteByte('0')
	}
	var buf [4]byte
	sb.Write(strconv.AppendInt(buf[:0], int64(v), 10))
}

// escapeText escapes parentheses and backslashes for PDF text strings.
func escapeText(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	return s
}
