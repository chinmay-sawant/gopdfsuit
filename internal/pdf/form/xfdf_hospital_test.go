package form

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func assertStartxrefPointsToXref(t *testing.T, pdf []byte) {
	t.Helper()

	sxIdx := bytes.LastIndex(pdf, []byte("startxref"))
	if sxIdx < 0 {
		t.Fatal("missing startxref")
	}

	rest := pdf[sxIdx+len("startxref"):]
	rest = bytes.TrimSpace(rest)
	lineEnd := bytes.IndexByte(rest, '\n')
	if lineEnd < 0 {
		t.Fatal("malformed startxref trailer")
	}
	offset, err := strconv.Atoi(string(bytes.TrimSpace(rest[:lineEnd])))
	if err != nil {
		t.Fatalf("parse startxref offset: %v", err)
	}
	if offset < 0 || offset >= len(pdf) {
		t.Fatalf("startxref offset %d out of range (len=%d)", offset, len(pdf))
	}

	at := pdf[offset:]
	if bytes.HasPrefix(at, []byte("xref\n")) || bytes.HasPrefix(at, []byte("xref\r\n")) {
		return
	}
	if bytes.Contains(at[:min(120, len(at))], []byte("/Type /XRef")) || bytes.Contains(at[:min(120, len(at))], []byte("/Type/XRef")) {
		return
	}
	t.Fatalf("startxref %d does not point to xref table or xref stream, got %q", offset, at[:min(40, len(at))])
}

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
	assertStartxrefPointsToXref(t, out)

	outPath := filepath.Join("..", "..", "..", "sampledata", "filler", "generated.pdf")
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		t.Fatalf("write generated.pdf: %v", err)
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
