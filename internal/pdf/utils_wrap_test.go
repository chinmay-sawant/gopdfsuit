package pdf

import (
	"reflect"
	"testing"
)

const wrapTestFontHelvetica = "Helvetica"

func TestWrapTextIntoMatchesWrapText(t *testing.T) {
	registry := NewFontRegistry()
	fontName := wrapTestFontHelvetica
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

	first := WrapTextInto(&ws, "hello world", wrapTestFontHelvetica, 12, 200, registry)
	if len(first) == 0 {
		t.Fatal("expected at least one line")
	}

	firstCap := cap(ws.buf)

	second := WrapTextInto(&ws, "another wrap test line", wrapTestFontHelvetica, 12, 40, registry)
	if len(second) == 0 {
		t.Fatal("expected at least one line from second call")
	}

	if cap(ws.buf) < firstCap {
		t.Fatalf("buffer capacity shrank: first=%d second=%d", firstCap, cap(ws.buf))
	}
}

func TestWrapTextIntoStaleSlicesWithoutClone(t *testing.T) {
	registry := NewFontRegistry()
	var ws WrapState

	email := WrapTextInto(&ws, "user1@example.com", wrapTestFontHelvetica, 10, 40, registry)
	_ = WrapTextInto(&ws, "Lorem ipsum dolor sit amet, consectetur adipiscing elit.", wrapTestFontHelvetica, 10, 50, registry)

	if len(email) == 0 {
		t.Fatal("expected wrapped email line")
	}
	got := string(email[0])
	if got == "user1@example.com" {
		t.Fatal("expected reused WrapState to invalidate prior slice views")
	}
}

func TestCloneWrapLinesSurvivesNextWrap(t *testing.T) {
	registry := NewFontRegistry()
	var ws WrapState

	email := cloneWrapLines(WrapTextInto(&ws, "user1@example.com", wrapTestFontHelvetica, 10, 200, registry))
	_ = WrapTextInto(&ws, "Lorem ipsum dolor sit amet, consectetur adipiscing elit.", wrapTestFontHelvetica, 10, 50, registry)

	if len(email) == 0 {
		t.Fatal("expected wrapped email line")
	}
	if got := string(email[0]); got != "user1@example.com" {
		t.Fatalf("cloned lines should stay stable: got %q", got)
	}
}
