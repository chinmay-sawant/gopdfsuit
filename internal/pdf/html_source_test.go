//go:build !js

package pdf

import "testing"

// TestHTMLSourceContentEmptyInputFailsFast pins the fail-fast adapter guard
// shared by the PDF and image paths. It lives in this !js-tagged file
// because htmlSourceContent is the server-only source policy; the js twin
// (inlineHTMLContent) is pinned in html_source_js_test.go.
func TestHTMLSourceContentEmptyInputFailsFast(t *testing.T) {
	if _, err := htmlSourceContent("", ""); err == nil {
		t.Fatal("expected error for empty source, got nil")
	}
}
