package form

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFillPDFWithXFDFCompressedSample(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "sampledata", "filler", "compressed")
	pdfPath := filepath.Join(baseDir, "medical_form.pdf")
	xfdfPath := filepath.Join(baseDir, "medical_data.xfdf")
	outPath := filepath.Join(baseDir, "generated.pdf")

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}

	xfdfBytes, err := os.ReadFile(xfdfPath)
	if err != nil {
		t.Fatalf("read xfdf: %v", err)
	}

	out, err := FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		t.Fatalf("fill compressed pdf: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("filled output is empty")
	}

	if err := os.WriteFile(outPath, out, 0644); err != nil {
		t.Fatalf("write generated.pdf: %v", err)
	}

	fields, err := DetectFormFieldsAdvanced(out)
	if err != nil {
		t.Fatalf("detect fields from output: %v", err)
	}

	if got := fields["patient_name"]; got != "John Smith" {
		t.Fatalf("patient_name mismatch: got %q", got)
	}

	notes := fields["doctor_notes"]
	if !strings.Contains(notes, "Prescribed antibiotics") {
		t.Fatalf("doctor_notes not filled correctly: got %q", notes)
	}
}
