package gopdflib_test

import (
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func TestMakePropsRoundtrip(t *testing.T) {
	got := gopdflib.MakeProps("Helvetica", 12, true, false, false, "left", [4]int{1, 1, 1, 1})
	if got != "Helvetica:12:100:left:1:1:1:1" {
		t.Fatalf("MakeProps = %q, want %q", got, "Helvetica:12:100:left:1:1:1:1")
	}

	opts := gopdflib.ParseFontOpts(got)
	if opts.Name != "Helvetica" || opts.Size != 12 {
		t.Fatalf("roundtrip name/size = %q/%d, want Helvetica/12", opts.Name, opts.Size)
	}
	if !opts.Bold || opts.Italic || opts.Underline {
		t.Fatalf("roundtrip style = %v/%v/%v, want true/false/false", opts.Bold, opts.Italic, opts.Underline)
	}
	if opts.Align != gopdflib.AlignLeft {
		t.Fatalf("roundtrip align = %q, want left", opts.Align)
	}
	if opts.Borders != gopdflib.Borders([4]int{1, 1, 1, 1}) {
		t.Fatalf("roundtrip borders = %v, want [1 1 1 1]", opts.Borders)
	}
	if opts.String() != got {
		t.Fatalf("String after parse = %q, want %q", opts.String(), got)
	}
}

func TestFontOptsValidation(t *testing.T) {
	// Unknown alignment falls back to left.
	if got := gopdflib.MakeProps("Helvetica", 12, false, false, false, "justify", [4]int{0, 0, 0, 0}); got != "Helvetica:12:000:left:0:0:0:0" {
		t.Fatalf("bad align = %q, want left fallback", got)
	}
	// Non-positive size falls back to 12.
	if got := gopdflib.MakeProps("Helvetica", 0, false, false, false, "center", [4]int{0, 0, 0, 0}); got != "Helvetica:12:000:center:0:0:0:0" {
		t.Fatalf("zero size = %q, want 12 fallback", got)
	}
	// Empty name falls back to Helvetica.
	if got := (gopdflib.FontOpts{Align: gopdflib.AlignRight}).String(); got != "Helvetica:12:000:right:0:0:0:0" {
		t.Fatalf("empty opts = %q, want Helvetica default", got)
	}
	// Short style codes are not 3 chars, so ParseFontOpts treats them as regular.
	parsed := gopdflib.ParseFontOpts("Helvetica:12:10:left:1:1:1:1")
	if parsed.Bold || parsed.Italic || parsed.Underline {
		t.Fatalf("short style code parsed as styled: %+v", parsed)
	}
	if got := parsed.String(); got != "Helvetica:12:000:left:1:1:1:1" {
		t.Fatalf("short style re-emit = %q, want 000", got)
	}
	// Unparseable size and unknown alignment fall back.
	parsed = gopdflib.ParseFontOpts("Courier:big:111:sideways:a:b:c:d")
	if parsed.Size != 12 || parsed.Align != gopdflib.AlignLeft {
		t.Fatalf("bad size/align = %d/%q, want 12/left", parsed.Size, parsed.Align)
	}
	if parsed.Borders != (gopdflib.Borders{}) {
		t.Fatalf("bad borders = %v, want zeros", parsed.Borders)
	}
}

func TestParseFontOptsStyles(t *testing.T) {
	parsed := gopdflib.ParseFontOpts("Times-Roman:10:011:right:0:1:0:1")
	if parsed.Name != "Times-Roman" || parsed.Size != 10 {
		t.Fatalf("name/size = %q/%d", parsed.Name, parsed.Size)
	}
	if parsed.Bold || !parsed.Italic || !parsed.Underline {
		t.Fatalf("style flags = %v/%v/%v, want false/true/true", parsed.Bold, parsed.Italic, parsed.Underline)
	}
	if parsed.Align != gopdflib.AlignRight {
		t.Fatalf("align = %q, want right", parsed.Align)
	}
	if parsed.Borders != gopdflib.Borders([4]int{0, 1, 0, 1}) {
		t.Fatalf("borders = %v", parsed.Borders)
	}
}
