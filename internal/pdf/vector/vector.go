// Package vector holds the single float, color, and text-escape policy for
// PDF content-stream emission shared by the SVG converter
// (internal/pdf/svg) and the Typst math renderer (typstsyntax).
//
// Both packages previously formatted floats, emitted fill/stroke colors,
// escaped literal strings, and stroked lines/paths with their own helpers.
// Routing them through this package keeps one policy: prec-parameterized
// 'f' formatting via strconv, backslash/paren escaping for literal strings,
// and "rg"/"RG" color operators with caller-chosen precision.
package vector

import (
	"bytes"
	"strconv"
	"strings"
)

// AppendFloat appends f formatted with prec fractional digits.
func AppendFloat(dst []byte, f float64, prec int) []byte {
	var scratch [64]byte
	return append(dst, strconv.AppendFloat(scratch[:0], f, 'f', prec, 64)...)
}

// WriteFloat writes one float with prec fractional digits.
func WriteFloat(b *bytes.Buffer, prec int, f float64) {
	var scratch [64]byte
	b.Write(strconv.AppendFloat(scratch[:0], f, 'f', prec, 64))
}

// WriteFloats writes space-separated floats with prec fractional digits.
func WriteFloats(b *bytes.Buffer, prec int, nums ...float64) {
	for i, n := range nums {
		if i > 0 {
			b.WriteByte(' ')
		}
		WriteFloat(b, prec, n)
	}
}

// FormatFloat formats f with 2 fractional digits for content streams.
func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// EscapeText escapes backslash and parentheses for PDF literal strings.
func EscapeText(s string) string {
	if !strings.ContainsAny(s, "\\()") {
		return s
	}
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// SetFill emits "r g b rg" with prec fractional digits.
func SetFill(b *bytes.Buffer, prec int, r, g, bl float64) {
	WriteFloats(b, prec, r, g, bl)
	b.WriteString(" rg\n")
}

// SetStroke emits "r g b RG" with prec fractional digits.
func SetStroke(b *bytes.Buffer, prec int, r, g, bl float64) {
	WriteFloats(b, prec, r, g, bl)
	b.WriteString(" RG\n")
}

// LineWidth emits "w w".
func LineWidth(b *bytes.Buffer, prec int, w float64) {
	WriteFloat(b, prec, w)
	b.WriteString(" w\n")
}

// StrokeLine strokes one segment: "x1 y1 m x2 y2 l S".
func StrokeLine(b *bytes.Buffer, prec int, x1, y1, x2, y2 float64) {
	WriteFloats(b, prec, x1, y1)
	b.WriteString(" m ")
	WriteFloats(b, prec, x2, y2)
	b.WriteString(" l S\n")
}

// StrokePath strokes a polyline through pts: moveto plus linetos plus S.
// Callers need at least 2 points; fewer is a no-op.
func StrokePath(b *bytes.Buffer, prec int, pts [][2]float64) {
	if len(pts) < 2 {
		return
	}
	WriteFloats(b, prec, pts[0][0], pts[0][1])
	b.WriteString(" m\n")
	for _, pt := range pts[1:] {
		WriteFloats(b, prec, pt[0], pt[1])
		b.WriteString(" l\n")
	}
	b.WriteString("S\n")
}
