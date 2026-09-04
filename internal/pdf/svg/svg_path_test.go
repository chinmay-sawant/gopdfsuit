package svg

import (
	"bytes"
	"strings"
	"testing"
)

// TestParsePathDataTruncated proves crafted/truncated d attributes abort the
// path instead of panicking with index out of range.
func TestParsePathDataTruncated(t *testing.T) {
	for _, d := range []string{
		"M 10",
		"M",
		"L 5",
		"C 1 2 3",
		"Q 1 2 3",
		"H",
		"M 10 20 L 30",
		"M xx yy",
		"",
	} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("parsePathData(%q) panicked: %v", d, r)
				}
			}()
			var b bytes.Buffer
			parsePathData(&b, d)
		}()
	}
}

// TestParsePathDataValid ensures the bounds-checked rewrite still emits the
// same operators for well-formed paths.
func TestParsePathDataValid(t *testing.T) {
	var b bytes.Buffer
	parsePathData(&b, "M 10 20 L 30 40 H 50 V 60 Z")
	out := b.String()
	for _, want := range []string{"m", "l", "h"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in %q", want, out)
		}
	}

	b.Reset()
	parsePathData(&b, "M 0 0 C 1 1 2 2 3 3 Q 4 4 5 5")
	if !strings.Contains(b.String(), "c") {
		t.Fatalf("expected curve operators in %q", b.String())
	}
}
