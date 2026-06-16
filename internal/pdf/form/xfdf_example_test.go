package form

import (
	_ "embed"
	"os"
	"testing"
)

// This test demonstrates DetectFormFields and FillPDFWithXFDF on the sample files
func TestXFDFFillSample(t *testing.T) {
	pdfBytes, err := os.ReadFile("../../../sampledata/oldata/pdf+xfdf/us_hospital_encounter_acroform.pdf")
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	xfdfBytes, err := os.ReadFile("../../../sampledata/oldata/pdf+xfdf/us_hospital_encounter_data.xfdf")
	if err != nil {
		t.Fatalf("read xfdf: %v", err)
	}

	// Test enhanced field detection
	detectedFieldsMap, err := DetectFormFieldsAdvanced(pdfBytes)
	if err != nil {
		t.Fatalf("detect fields advanced: %v", err)
	}
	if len(detectedFieldsMap) == 0 {
		t.Logf("no fields detected with advanced detection")
	} else {
		t.Logf("detected fields with values: %v", detectedFieldsMap)
	}

	// Test original field detection for comparison
	fields, err := DetectFormFields(pdfBytes)
	if err != nil {
		t.Fatalf("detect fields: %v", err)
	}
	if len(fields) == 0 {
		t.Logf("no fields detected (heuristic)")
	} else {
		t.Logf("detected field names: %v", fields)
	}

	// // Test enhanced filling
	// out, err := FillPDFWithXFDFAdvanced(pdfBytes, xfdfBytes)
	// if err != nil {
	// 	t.Fatalf("fill advanced: %v", err)
	// }
	// if len(out) == 0 {
	// 	t.Fatalf("output empty")
	// }
	// _ = os.WriteFile("filled_sample_advanced.pdf", out, 0644)

	out, err := FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		t.Fatalf("fill: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("output empty")
	}
	_ = os.WriteFile("../../../sampledata/oldata/pdf+xfdf/filled_sample.pdf", out, 0644)

	expected, err := ParseXFDF(xfdfBytes)
	if err != nil {
		t.Fatalf("parse xfdf: %v", err)
	}
	filled, err := DetectFormFieldsAdvanced(out)
	if err != nil {
		t.Fatalf("detect filled fields: %v", err)
	}
	for name, want := range expected {
		if got := filled[name]; got != want {
			t.Errorf("field %q: got %q want %q", name, got, want)
		}
	}
}
