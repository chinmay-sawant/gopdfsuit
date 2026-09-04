package svg

import (
	"strings"
	"testing"
)

// Deterministic fuzz seeds as plain unit tests: truncated and malformed SVG
// inputs must never panic. They may return an error or best-effort output.
func TestConvertSVGTruncatedSeedsNoPanic(t *testing.T) {
	seeds := []string{
		``,
		`<`,
		`<svg`,
		`<svg>`,
		`<svg width="`,
		`<svg width="100" height="100"><rect`,
		`<svg width="100" height="100"><rect x="10" y=`,
		`<svg width="100" height="100"><path d="M10 10 L20`,
		"<svg><path d=\"M10 10 C 20 20, 40 40",
		"<svg viewBox=\"0 0",
		"<svg width=\"abc\" height=\"def\"><circle cx=\"50\" cy=\"50\" r=\"",
		"not xml at all {{{{",
		"<svg xmlns=\"http://www.w3.org/2000/svg\"><g><g><g>",
		"<svg><text x=\"10\">unclosed text",
		"<svg><polygon points=\"10,10 20,20 30",
		"<svg><polyline points=\"",
		"<svg><ellipse cx=\"50\" cy=\"50\" rx=\"30\" ry=\"",
		"<svg><line x1=\"0\" y1=\"0\" x2=\"",
		"<svg width=\"1e999999\" height=\"-5\"><rect width=\"NaN\"/>",
		strings.Repeat("<svg><g>", 500),
		strings.Repeat("M10 10 L20 20 ", 2000),
		"<svg><path d=\"" + strings.Repeat("M", 5000),
		"<svg><!-- unterminated comment",
		"<svg><![CDATA[unterminated cdata",
		"<svg><rect " + strings.Repeat("x=\"1\" ", 3000) + "/>",
	}
	for i, s := range seeds {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("seed %d panicked: %v (input %q)", i, r, truncate(s))
				}
			}()
			_, _, _, _ = ConvertSVGToPDFCommands([]byte(s))
		}()
	}
}

func TestConvertSVGWellFormedStillWorks(t *testing.T) {
	data := []byte(`<svg width="100" height="100"><rect x="10" y="10" width="50" height="50"/></svg>`)
	cmds, w, h, err := ConvertSVGToPDFCommands(data)
	if err != nil {
		t.Fatalf("valid SVG errored: %v", err)
	}
	if w != 100 || h != 100 {
		t.Fatalf("unexpected dimensions %dx%d", w, h)
	}
	if len(cmds) == 0 {
		t.Fatal("expected non-empty PDF commands for valid SVG")
	}
}

func truncate(s string) string {
	if len(s) > 80 {
		return s[:80] + "..."
	}
	return s
}

// svgFuzzSeeds mirrors the deterministic regression seeds above so the fuzzer
// starts from known-tricky shapes (truncation, malformed numbers, deep
// nesting, unterminated constructs).
var svgFuzzSeeds = []string{
	``,
	`<`,
	`<svg`,
	`<svg>`,
	`<svg width="`,
	`<svg width="100" height="100"><rect`,
	`<svg width="100" height="100"><rect x="10" y=`,
	`<svg width="100" height="100"><path d="M10 10 L20`,
	"<svg><path d=\"M10 10 C 20 20, 40 40",
	"<svg viewBox=\"0 0",
	"<svg width=\"abc\" height=\"def\"><circle cx=\"50\" cy=\"50\" r=\"",
	"not xml at all {{{{",
	"<svg xmlns=\"http://www.w3.org/2000/svg\"><g><g><g>",
	"<svg><text x=\"10\">unclosed text",
	"<svg><polygon points=\"10,10 20,20 30",
	"<svg><polyline points=\"",
	"<svg><ellipse cx=\"50\" cy=\"50\" rx=\"30\" ry=\"",
	"<svg><line x1=\"0\" y1=\"0\" x2=\"",
	"<svg width=\"1e999999\" height=\"-5\"><rect width=\"NaN\"/>",
	"<svg><!-- unterminated comment",
	"<svg><![CDATA[unterminated cdata",
	`<svg width="100" height="100"><rect x="10" y="10" width="50" height="50"/></svg>`,
	`<svg viewBox="0 0 100 100"><circle cx="50" cy="50" r="40"/></svg>`,
	`<svg><path d="M10 10 L90 90 M20 20 C30 30 40 40 50 50 Z"/></svg>`,
	`<svg><g transform="translate(10, 20) scale(2)"><rect width="5" height="5"/></g></svg>`,
	`<svg><defs><rect id="r" width="5" height="5"/></defs><use href="#r" x="1" y="2"/></svg>`,
}

// FuzzConvertSVG asserts truncated and malformed SVG inputs never panic.
// They may return an error or best-effort output. Seed corpus is committed
// under testdata/fuzz/FuzzConvertSVG so plain `go test` replays it.
func FuzzConvertSVG(f *testing.F) {
	for _, s := range svgFuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, data string) {
		_, _, _, _ = ConvertSVGToPDFCommands([]byte(data))
	})
}
