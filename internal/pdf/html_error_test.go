package pdf

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// Empty requests must fail fast in the validation layer before the
// gowkhtmltopdf engine is ever touched, so these tests need no browser and
// never skip.
func TestConvertHTMLToPDFEmptyInputFailsFast(t *testing.T) {
	_, err := ConvertHTMLToPDF(models.HTMLToPDFRequest{})
	if err == nil {
		t.Fatal("expected error for empty HTMLToPDFRequest, got nil")
	}
	if !strings.Contains(err.Error(), "either HTML content or URL must be provided") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConvertHTMLToImageEmptyInputFailsFast(t *testing.T) {
	_, err := ConvertHTMLToImage(models.HTMLToImageRequest{})
	if err == nil {
		t.Fatal("expected error for empty HTMLToImageRequest, got nil")
	}
	if !strings.Contains(err.Error(), "either HTML content or URL must be provided") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestHTMLToImageSVGUnsupported pins the gowkhtmltopdf gap: svg has no
// engine equivalent and must fail fast without rendering.
func TestHTMLToImageSVGUnsupported(t *testing.T) {
	_, err := ConvertHTMLToImage(models.HTMLToImageRequest{
		HTML:   "<html><body><h1>Hello</h1></body></html>",
		Format: "svg",
	})
	if err == nil {
		t.Fatal("expected error for svg format, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unsupported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestHTMLMarginParsing pins the "10mm"-style margin mapping onto mm floats.
func TestHTMLMarginParsing(t *testing.T) {
	cases := map[string]float64{
		"":      10,
		"10mm":  10,
		"1cm":   10,
		"1in":   25.4,
		"72pt":  25.4,
		"10":    10,
		"bogus": 10,
	}
	for raw, want := range cases {
		got := parseMarginMM(raw)
		if got < want-1e-9 || got > want+1e-9 {
			t.Errorf("parseMarginMM(%q) = %v, want %v", raw, got, want)
		}
	}
}

// TestHTMLSourceContentEmptyInputFailsFast pins the fail-fast adapter guard
// shared by the PDF and image paths.
func TestHTMLSourceContentEmptyInputFailsFast(t *testing.T) {
	if _, err := htmlSourceContent("", ""); err == nil {
		t.Fatal("expected error for empty source, got nil")
	}
}
