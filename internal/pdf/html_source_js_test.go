//go:build js

package pdf

import (
	"strings"
	"testing"
)

// TestInlineHTMLContentEmptyInputFailsFast pins the fail-fast adapter guard
// shared by the PDF and image paths in the browser build.
func TestInlineHTMLContentEmptyInputFailsFast(t *testing.T) {
	if _, err := inlineHTMLContent("", ""); err == nil {
		t.Fatal("expected error for empty source, got nil")
	}
}

// TestInlineHTMLContentURLOnlyFailsFast pins the js-build source policy:
// the engine loader cannot dial under js/wasm, so URL-only input fails fast
// with guidance (the binding layer pre-fetches page HTML via browser fetch)
// instead of surfacing a raw DNS dial error.
func TestInlineHTMLContentURLOnlyFailsFast(t *testing.T) {
	_, err := inlineHTMLContent("", "https://example.com")
	if err == nil {
		t.Fatal("expected error for URL-only source, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "browser fetch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestInlineHTMLContentKeepsBaseURL pins that HTML plus a page URL becomes
// an inline document carrying the page URL as base, so relative
// subresources resolve against the page origin.
func TestInlineHTMLContentKeepsBaseURL(t *testing.T) {
	content, err := inlineHTMLContent("<html></html>", "https://example.com/page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(content.HTML) == 0 {
		t.Fatal("expected inline HTML content, got empty")
	}
	if content.Base != "https://example.com/page" {
		t.Fatalf("unexpected base URL: %q", content.Base)
	}
	if content.URL != "" {
		t.Fatalf("expected no engine-fetch URL, got %q", content.URL)
	}
}
