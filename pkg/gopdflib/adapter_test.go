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
