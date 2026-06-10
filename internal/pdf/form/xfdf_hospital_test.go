package form

import (
	"os"
	"testing"
)

func TestHospitalFillFieldValues(t *testing.T) {
	pdfBytes, err := os.ReadFile("../../../sampledata/filler/us_hospital_encounter_acroform.pdf")
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	xfdfBytes, err := os.ReadFile("../../../sampledata/filler/us_hospital_encounter_data.xfdf")
	if err != nil {
		t.Fatalf("read xfdf: %v", err)
	}

	expected, err := ParseXFDF(xfdfBytes)
	if err != nil {
		t.Fatalf("parse xfdf: %v", err)
	}

	out, err := FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		t.Fatalf("fill: %v", err)
	}

	got, err := DetectFormFieldsAdvanced(out)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}

	for name, want := range expected {
		if have := got[name]; have != want {
			t.Errorf("field %q: got %q want %q", name, have, want)
		}
	}
}