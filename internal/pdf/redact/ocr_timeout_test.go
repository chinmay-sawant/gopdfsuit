package redact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

// TestOCRExtractWordsTimeout proves a hanging pdftoppm cannot block the
// pipeline: a stub binary that sleeps must be killed at ocrCommandTimeout.
func TestOCRExtractWordsTimeout(t *testing.T) {
	stubDir := t.TempDir()
	stub := "#!/bin/sh\nsleep 30\n"
	for _, name := range []string{"pdftoppm", "tesseract"} {
		p := filepath.Join(stubDir, name)
		if err := os.WriteFile(p, []byte(stub), 0o755); err != nil {
			t.Fatalf("write stub %s: %v", name, err)
		}
	}
	t.Setenv("PATH", stubDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	old := ocrCommandTimeout
	ocrCommandTimeout = time.Second
	defer func() { ocrCommandTimeout = old }()

	start := time.Now()
	_, err := tesseractProvider{}.ExtractWords(minimalPDF, models.OCRSettings{Provider: "tesseract"})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error from hanging pdftoppm, got nil")
	}
	if elapsed > 15*time.Second {
		t.Fatalf("OCR call took %v, timeout did not fire", elapsed)
	}
	if !strings.Contains(err.Error(), "pdftoppm") {
		t.Fatalf("expected pdftoppm error, got: %v", err)
	}
}

func TestOCRProviderRejectsUnknown(t *testing.T) {
	if _, err := getOCRProvider(models.OCRSettings{Provider: "nope"}); err == nil {
		t.Fatal("expected error for unknown OCR provider")
	}
}
