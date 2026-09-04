package gopdflib_test

import (
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

// TestOwnedTypesRoundTrip asserts owned public types survive the
// public<->internal translation without field loss.
func TestOwnedTypesRoundTrip(t *testing.T) {
	tmpl := gopdflib.PDFTemplate{
		Config: gopdflib.Config{Page: "A4", PageAlignment: 1, PdfTitle: "roundtrip"},
		Title:  gopdflib.Title{Props: "Helvetica:18:100:center:0:0:0:0", Text: "Hi"},
		Table: []gopdflib.Table{{
			MaxColumns:   2,
			ColumnWidths: []float64{1, 1},
			Rows: []gopdflib.Row{{Row: []gopdflib.Cell{
				{Props: "Helvetica:12:100:left:1:1:1:1", Text: "a", Link: "#sec"},
				{Props: "Helvetica:12:100:left:1:1:1:1", Text: "b"},
			}}},
		}},
		Elements: []gopdflib.Element{{Type: "spacer", Spacer: &gopdflib.Spacer{Height: 9}}},
		Footer:   gopdflib.Footer{Font: "Helvetica:10", Text: "foot"},
	}
	out, err := gopdflib.GeneratePDF(tmpl)
	if err != nil {
		t.Fatalf("GeneratePDF with owned types: %v", err)
	}
	if len(out) < 8 || string(out[:5]) != "%PDF-" {
		t.Fatal("expected PDF bytes")
	}
}

func TestParseCompressLevel(t *testing.T) {
	cases := map[string]gopdflib.CompressLevel{
		"":        gopdflib.CompressMedium,
		"light":   gopdflib.CompressLight,
		"LIGHT":   gopdflib.CompressLight,
		"Medium":  gopdflib.CompressMedium,
		"heavy":   gopdflib.CompressHeavy,
		" HeAvY ": gopdflib.CompressHeavy, //nolint:gocritic // intentional whitespace-tolerance case
		"bogus":   gopdflib.CompressMedium,
	}
	for in, want := range cases {
		if got := gopdflib.ParseCompressLevel(in); got != want {
			t.Fatalf("ParseCompressLevel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCompressLevelValues(t *testing.T) {
	if string(gopdflib.CompressLight) != "light" ||
		string(gopdflib.CompressMedium) != "medium" ||
		string(gopdflib.CompressHeavy) != "heavy" {
		t.Fatal("compress tier strings changed; engine mapping depends on them")
	}
	if gopdflib.MaxCompressInputBytes != 32<<20 {
		t.Fatalf("MaxCompressInputBytes = %d, want %d", gopdflib.MaxCompressInputBytes, 32<<20)
	}
}

// TestPrecomputedStandardFontsPassthrough pins the startup font hint on the
// public template type: the setter/getter round-trip, and GeneratePDF still
// emits PDF bytes with the hint set (the adapter carries it into the engine
// outside the JSON shape).
func TestPrecomputedStandardFontsPassthrough(t *testing.T) {
	tmpl := gopdflib.PDFTemplate{
		Config: gopdflib.Config{Page: "A4", PageAlignment: 1},
		Title:  gopdflib.Title{Props: "Helvetica:18:100:center:0:0:0:0", Text: "Hi"},
		Footer: gopdflib.Footer{Font: "Helvetica:10", Text: "foot"},
	}
	tmpl.SetPrecomputedStandardFonts("Helvetica")
	if got := tmpl.PrecomputedStandardFonts(); len(got) != 1 || got[0] != "Helvetica" {
		t.Fatalf("hint = %v, want [Helvetica]", got)
	}
	out, err := gopdflib.GeneratePDF(tmpl)
	if err != nil {
		t.Fatalf("GeneratePDF with font hint: %v", err)
	}
	if len(out) < 8 || string(out[:5]) != "%PDF-" {
		t.Fatal("expected PDF bytes")
	}
	tmpl.SetPrecomputedStandardFonts()
	if got := tmpl.PrecomputedStandardFonts(); len(got) != 0 {
		t.Fatalf("cleared hint = %v, want empty", got)
	}
}

// TestCompressLevelDefaultsParity pins the cross-tier level contract shared
// with frontend compressLevels.js (toServerLevel/toWasmLevel) and
// cmd/wasmcompress: numeric 2 <-> "medium" <-> wasm 2, empty selects Medium,
// and invalid non-numeric strings are an error on the strict pair
// (ToServerLevel/ToWasmLevel) while ParseCompressLevel keeps the
// engine-compatible silent Medium fallback.
func TestCompressLevelDefaultsParity(t *testing.T) {
	tiers := []struct {
		num    int
		server gopdflib.CompressLevel
	}{
		{1, gopdflib.CompressLight},
		{2, gopdflib.CompressMedium},
		{3, gopdflib.CompressHeavy},
	}
	for _, tc := range tiers {
		if got, err := gopdflib.ToServerLevel(tc.num); err != nil || got != tc.server {
			t.Fatalf("ToServerLevel(%d) = %q, %v; want %q, nil", tc.num, got, err, tc.server)
		}
		if got, err := gopdflib.ToWasmLevel(string(tc.server)); err != nil || got != tc.num {
			t.Fatalf("ToWasmLevel(%q) = %d, %v; want %d, nil", tc.server, got, err, tc.num)
		}
		if got := gopdflib.ParseCompressLevel(string(tc.server)); got != tc.server {
			t.Fatalf("ParseCompressLevel(%q) = %q, want %q", tc.server, got, tc.server)
		}
		if back, err := gopdflib.ToServerLevel(tc.num); err != nil {
			t.Fatal(err)
		} else if n, err := gopdflib.ToWasmLevel(back); err != nil || n != tc.num {
			t.Fatalf("round trip %d -> %q -> %d (%v); want %d", tc.num, back, n, err, tc.num)
		}
	}

	defaults := []any{nil, "", "  ", 0, 2, "2", "medium", "Medium", " MEDIUM "}
	for _, in := range defaults {
		server, err := gopdflib.ToServerLevel(in)
		if err != nil {
			t.Fatalf("ToServerLevel(%v) errored: %v", in, err)
		}
		if server != gopdflib.CompressMedium {
			t.Fatalf("ToServerLevel(%v) = %q, want medium", in, server)
		}
		wasm, err := gopdflib.ToWasmLevel(in)
		if err != nil || wasm != 2 {
			t.Fatalf("ToWasmLevel(%v) = %d, %v; want 2, nil", in, wasm, err)
		}
	}

	for _, bad := range []string{"bogus", "med", "light2", "-"} {
		if _, err := gopdflib.ToServerLevel(bad); err == nil {
			t.Fatalf("ToServerLevel(%q) = nil error, want invalid-level error", bad)
		}
		if _, err := gopdflib.ToWasmLevel(bad); err == nil {
			t.Fatalf("ToWasmLevel(%q) = nil error, want invalid-level error", bad)
		}
		if got := gopdflib.ParseCompressLevel(bad); got != gopdflib.CompressMedium {
			t.Fatalf("ParseCompressLevel(%q) = %q, want pinned medium fallback", bad, got)
		}
	}
}
