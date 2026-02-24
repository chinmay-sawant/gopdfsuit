package form

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJeffersonFill(t *testing.T) {
	baseDir := "../../../sampledata/filler/jefferson"
	pdfPath := filepath.Join(baseDir, "jefferson.pdf")
	xfdfPath := filepath.Join(baseDir, "jefferson.xfdf")
	outPath := filepath.Join(baseDir, "jefferson_filled.pdf")

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}

	xfdfBytes, err := os.ReadFile(xfdfPath)
	if err != nil {
		t.Fatalf("read xfdf: %v", err)
	}

	// Test filling
	out, err := FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		t.Fatalf("fill jefferson pdf: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("filled output is empty")
	}

	if err := os.WriteFile(outPath, out, 0644); err != nil {
		t.Fatalf("write jefferson_filled.pdf: %v", err)
	}

	// Verify fields
	fields, err := DetectFormFieldsAdvanced(out)
	if err != nil {
		// If advanced detection fails, try naive
		fields, err = detectFormFieldsNaive(out)
		if err != nil {
			t.Fatalf("detect fields from output: %v", err)
		}
	}

	expectedFields := map[string]string{
		"Patient Name":  "John Doe",
		"MRN":           "123456",
		"Date of Birth": "1/1/1982",
		"Provider Name": "Patricia Samson",
	}

	for k, v := range expectedFields {
		if got := fields[k]; got != v {
			t.Errorf("field %q mismatch: got %q, want %q", k, got, v)
		}
	}
}
