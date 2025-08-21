package pdf

import (
	_ "embed"
	"os"
	"testing"
)

// This test demonstrates DetectFormFields and FillPDFWithXFDF on the sample files
func TestXFDFFillSample(t *testing.T) {
	pdfBytes, err := os.ReadFile("../..//sampledata/patientreg/patientreg.pdf")
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	xfdfBytes, err := os.ReadFile("../..//sampledata/patientreg/patientreg.xfdf")
	if err != nil {
		t.Fatalf("read xfdf: %v", err)
	}

	fields, err := DetectFormFields(pdfBytes)
	if err != nil {
		t.Fatalf("detect fields: %v", err)
	}
	if len(fields) == 0 {
		t.Logf("no fields detected (heuristic)")
	} else {
		t.Logf("detected fields: %v", fields)
	}

	out, err := FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		t.Fatalf("fill: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("output empty")
	}
	// write out for manual inspection
	_ = os.WriteFile("filled_sample.pdf", out, 0644)
}
