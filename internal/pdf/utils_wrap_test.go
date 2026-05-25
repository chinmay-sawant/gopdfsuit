package pdf

import (
	"reflect"
	"testing"
)

func TestWrapTextIntoMatchesWrapText(t *testing.T) {
	registry := NewFontRegistry()
	fontName := "Helvetica"
	fontSize := 12.0

	tests := []struct {
		name     string
		text     string
		maxWidth float64
	}{
		{
			name:     "empty text",
			text:     "",
			maxWidth: 100,
		},
		{
			name:     "single word",
			text:     "Hello",
			maxWidth: 200,
		},
		{
			name:     "multiple words wrap",
			text:     "The quick brown fox jumps over the lazy dog",
			maxWidth: 60,
		},
		{
			name:     "long word breaks",
			text:     "Supercalifragilisticexpialidocious",
			maxWidth: 40,
		},
		{
			name:     "whitespace only",
			text:     "   \t\n  ",
			maxWidth: 100,
		},
		{
			name:     "zero max width",
			text:     "fallback line",
			maxWidth: 0,
		},
		{
			name:     "mixed spacing",
			text:     "  word1   word2  word3  ",
			maxWidth: 50,
		},
	}

	var ws WrapState
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := WrapText(tt.text, fontName, fontSize, tt.maxWidth, registry)

			byteLines := WrapTextInto(&ws, tt.text, fontName, fontSize, tt.maxWidth, registry)
			got := make([]string, len(byteLines))
			for i, line := range byteLines {
				got[i] = string(line)
			}

			if !reflect.DeepEqual(want, got) {
				t.Fatalf("WrapTextInto mismatch:\nwant: %q\ngot:  %q", want, got)
			}
		})
	}
}

func TestWrapTextIntoReusesBuffers(t *testing.T) {
	registry := NewFontRegistry()
	var ws WrapState

	first := WrapTextInto(&ws, "hello world", "Helvetica", 12, 200, registry)
	if len(first) == 0 {
		t.Fatal("expected at least one line")
	}

	firstCap := cap(ws.buf)

	second := WrapTextInto(&ws, "another wrap test line", "Helvetica", 12, 40, registry)
	if len(second) == 0 {
		t.Fatal("expected at least one line from second call")
	}

	if cap(ws.buf) < firstCap {
		t.Fatalf("buffer capacity shrank: first=%d second=%d", firstCap, cap(ws.buf))
	}
}
