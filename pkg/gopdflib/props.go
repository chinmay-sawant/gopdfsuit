package gopdflib

import (
	"strconv"
	"strings"
)

// Align is a text alignment name used in a props string.
type Align string

// Supported alignment values. Anything else falls back to AlignLeft.
const (
	AlignLeft   Align = "left"
	AlignCenter Align = "center"
	AlignRight  Align = "right"
)

// normalizeAlign maps any input to a valid Align, falling back to left.
func normalizeAlign(s string) Align {
	switch Align(strings.ToLower(strings.TrimSpace(s))) {
	case AlignCenter:
		return AlignCenter
	case AlignRight:
		return AlignRight
	default:
		return AlignLeft
	}
}

// Color is a hex color string (for example "#B00020") carried in a cell's
// BgColor or TextColor field. Color never lives in the props grammar: the
// 3-digit style code means bold/italic/underline, not color.
type Color string

// Borders holds left, right, top, and bottom border flags in props order.
type Borders [4]int

// FontOpts is the typed form of a props string:
//
//	FontName:FontSize:StyleCode:Alignment:Left:Right:Top:Bottom
//
// For example "Helvetica:12:100:left:1:1:1:1". It imports no internal
// packages so snippet and frontend code can share the grammar without
// pulling in the engine.
type FontOpts struct {
	Name      string
	Size      int
	Bold      bool
	Italic    bool
	Underline bool
	Align     Align
	Borders   Borders
}

// styleCode renders the 3-character bold/italic/underline code.
func (f FontOpts) styleCode() string {
	var b [3]byte
	b[0] = '0'
	b[1] = '0'
	b[2] = '0'
	if f.Bold {
		b[0] = '1'
	}
	if f.Italic {
		b[1] = '1'
	}
	if f.Underline {
		b[2] = '1'
	}
	return string(b[:])
}

// String renders the props string. An empty name falls back to Helvetica,
// a size <= 0 falls back to 12, and an unknown alignment falls back to
// left, matching the engine's parseProps defaults.
func (f FontOpts) String() string {
	name := strings.TrimSpace(f.Name)
	if name == "" {
		name = "Helvetica"
	}
	size := f.Size
	if size <= 0 {
		size = 12
	}
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(size))
	sb.WriteByte(':')
	sb.WriteString(f.styleCode())
	sb.WriteByte(':')
	sb.WriteString(string(normalizeAlign(string(f.Align))))
	for _, v := range f.Borders {
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(v))
	}
	return sb.String()
}

// ParseFontOpts parses a props string into typed options. Only a
// 3-character style code is honored (anything else means regular); an
// unparseable or non-positive size falls back to 12, an unknown alignment
// falls back to left, and unparseable borders stay 0. This mirrors the
// engine's parseProps policy so overlay output parses back identically.
func ParseFontOpts(s string) FontOpts {
	out := FontOpts{Name: "Helvetica", Size: 12, Align: AlignLeft}
	parts := strings.Split(s, ":")
	if name := strings.TrimSpace(partAt(parts, 0)); name != "" {
		out.Name = name
	}
	if sizeStr := strings.TrimSpace(partAt(parts, 1)); sizeStr != "" {
		if n, err := strconv.Atoi(sizeStr); err == nil && n > 0 {
			out.Size = n
		}
	}
	if sc := partAt(parts, 2); len(sc) == 3 {
		out.Bold = sc[0] == '1'
		out.Italic = sc[1] == '1'
		out.Underline = sc[2] == '1'
	}
	out.Align = normalizeAlign(partAt(parts, 3))
	for i := 0; i < 4; i++ {
		if v, err := strconv.Atoi(strings.TrimSpace(partAt(parts, 4+i))); err == nil {
			out.Borders[i] = v
		}
	}
	return out
}

func partAt(parts []string, idx int) string {
	if idx < len(parts) {
		return parts[idx]
	}
	return ""
}
