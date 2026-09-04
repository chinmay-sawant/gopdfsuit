package pdf

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// Empty requests must fail fast in the validation layer before Chrome is
// ever touched, so these tests need no Chrome binary and never skip.
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
