package vector

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteFloatsPrec(t *testing.T) {
	var b bytes.Buffer
	WriteFloats(&b, 2, 1, 2.5, -0.125)
	if got, want := b.String(), "1.00 2.50 -0.12"; got != want {
		t.Fatalf("WriteFloats = %q, want %q", got, want)
	}
}

func TestFormatFloat(t *testing.T) {
	if got := FormatFloat(3.14159); got != "3.14" {
		t.Fatalf("FormatFloat = %q, want 3.14", got)
	}
}

func TestEscapeText(t *testing.T) {
	if got := EscapeText("a(b)\\c"); got != `a\(b\)\\c` {
		t.Fatalf("EscapeText = %q", got)
	}
	if got := EscapeText("plain"); got != "plain" {
		t.Fatalf("EscapeText fast path = %q", got)
	}
}

func TestSetFillStroke(t *testing.T) {
	var b bytes.Buffer
	SetFill(&b, 2, 1, 0, 0)
	SetStroke(&b, 3, 0, 0.5, 1)
	out := b.String()
	if !strings.Contains(out, "1.00 0.00 0.00 rg\n") {
		t.Fatalf("missing fill op in %q", out)
	}
	if !strings.Contains(out, "0.000 0.500 1.000 RG\n") {
		t.Fatalf("missing stroke op in %q", out)
	}
}

func TestStrokeLinePath(t *testing.T) {
	var b bytes.Buffer
	LineWidth(&b, 2, 0.5)
	StrokeLine(&b, 2, 0, 0, 10, 5)
	out := b.String()
	if !strings.Contains(out, "0.50 w\n") || !strings.Contains(out, " m ") || !strings.Contains(out, " l S\n") {
		t.Fatalf("line ops = %q", out)
	}
	b.Reset()
	StrokePath(&b, 2, [][2]float64{{0, 0}, {1, 1}, {2, 0}})
	out = b.String()
	if !strings.Contains(out, " m\n") || !strings.Contains(out, " l\nS\n") {
		t.Fatalf("path ops = %q", out)
	}
	b.Reset()
	StrokePath(&b, 2, [][2]float64{{0, 0}})
	if b.Len() != 0 {
		t.Fatalf("short path should be a no-op, got %q", b.String())
	}
}
