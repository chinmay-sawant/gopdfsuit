package pdf

import (
	"strconv"
	"testing"
)

// Canonical props-grammar pins for parseProps.
//
// Policy (shared with pkg/gopdflib ParseFontOpts / FontOpts.String, which
// cannot be imported here without an import cycle, so equality is pinned
// through identical literals on both sides):
//   - empty name falls back to Helvetica
//   - size <= 0 or unparseable falls back to 12
//   - only a 3-character style code is honored, else regular "000"
//   - unknown alignment falls back to left (center/right kept, any case)
func TestParsePropsCanonicalFallbacks(t *testing.T) {
	p := parseProps("")
	if p.FontName != "Helvetica" || p.FontSize != 12 || p.StyleCode != "000" {
		t.Fatalf("empty props = %+v, want Helvetica/12/000", p)
	}
	if p.Alignment != "left" || p.Bold || p.Italic || p.Underline {
		t.Fatalf("empty props = %+v, want left regular", p)
	}
	if p.Borders != [4]int{0, 0, 0, 0} {
		t.Fatalf("empty borders = %v, want zeros", p.Borders)
	}
}

func TestParsePropsCanonicalRenders(t *testing.T) {
	render := func(p string) string {
		parsed := parseProps(p)
		out := parsed.FontName + ":" + strconv.Itoa(parsed.FontSize) + ":" +
			parsed.StyleCode + ":" + parsed.Alignment
		for _, b := range parsed.Borders {
			out += ":" + strconv.Itoa(b)
		}
		return out
	}
	cases := []struct{ in, want string }{
		{"", "Helvetica:12:000:left:0:0:0:0"},
		{":0:000:left:0:0:0:0", "Helvetica:12:000:left:0:0:0:0"},
		{"Helvetica:0:100:center:1:1:1:1", "Helvetica:12:100:center:1:1:1:1"},
		{"Helvetica:12:100:justify:1:1:1:1", "Helvetica:12:100:left:1:1:1:1"},
		{"Helvetica:12:100:CENTER:1:1:1:1", "Helvetica:12:100:center:1:1:1:1"},
		{"Helvetica:12:10:left:1:1:1:1", "Helvetica:12:000:left:1:1:1:1"},
		{"Helvetica:abc:000:left:x:x:x:x", "Helvetica:12:000:left:0:0:0:0"},
	}
	for _, tc := range cases {
		if got := render(tc.in); got != tc.want {
			t.Errorf("parseProps(%q) renders %q, want %q", tc.in, got, tc.want)
		}
	}
}
