package gopdflib_test

// Phase 2.2 WASM per-op bindings: generate/merge/split entry points that the
// JS shim calls with serialized objects. These tests exercise the portable
// (non-js) surface against sampledata fixtures only.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func readWasmFixture(t *testing.T, elems ...string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(elems...))
	if err != nil {
		t.Fatalf("failed to read fixture %v: %v", elems, err)
	}
	return data
}

func assertPDF(t *testing.T, data []byte) {
	t.Helper()
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("output does not start with PDF header")
	}
	if !bytes.HasSuffix(data, []byte("%%EOF\n")) {
		t.Error("output does not end with EOF marker")
	}
}

func TestGeneratePDFFromJSONFinancialReport(t *testing.T) {
	raw := readWasmFixture(t, "..", "..", "sampledata", "financialreport", "financial_report.json")
	out, err := gopdflib.GeneratePDFFromJSON(raw)
	if err != nil {
		t.Fatalf("GeneratePDFFromJSON failed: %v", err)
	}
	assertPDF(t, out)
	t.Logf("generated PDF: %d bytes from %d bytes of template JSON", len(out), len(raw))
}

func TestGeneratePDFBorrowedFromJSONFinancialReport(t *testing.T) {
	raw := readWasmFixture(t, "..", "..", "sampledata", "financialreport", "financial_report.json")
	doc, err := gopdflib.GeneratePDFBorrowedFromJSON(raw)
	if err != nil {
		t.Fatalf("GeneratePDFBorrowedFromJSON failed: %v", err)
	}
	defer doc.Release()
	assertPDF(t, doc.CopyBytes())
}

func TestGeneratePDFFromJSONRejectsBadInput(t *testing.T) {
	if _, err := gopdflib.GeneratePDFFromJSON(nil); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Errorf("empty JSON: expected ErrInvalidInput, got %v", err)
	}
	if _, err := gopdflib.GeneratePDFFromJSON([]byte("{not json")); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Errorf("malformed JSON: expected ErrInvalidInput, got %v", err)
	}
	over := make([]byte, gopdflib.MaxTemplateJSONBytes+1)
	for i := range over {
		over[i] = ' '
	}
	if _, err := gopdflib.GeneratePDFFromJSON(over); !errors.Is(err, gopdflib.ErrLimitExceeded) {
		t.Errorf("oversize JSON: expected ErrLimitExceeded, got %v", err)
	}
}

func TestMergePDFsSampledataFixtures(t *testing.T) {
	a := readWasmFixture(t, "..", "..", "sampledata", "merge", "em-16.pdf")
	b := readWasmFixture(t, "..", "..", "sampledata", "merge", "em-19.pdf")
	merged, err := gopdflib.MergePDFs([][]byte{a, b})
	if err != nil {
		t.Fatalf("MergePDFs failed: %v", err)
	}
	assertPDF(t, merged)
	t.Logf("merged PDF: %d bytes from %d + %d", len(merged), len(a), len(b))
}

func TestSplitPDFWithSpecJSONPagesObject(t *testing.T) {
	src := readWasmFixture(t, "..", "..", "sampledata", "merge", "em-16.pdf")

	parts, err := gopdflib.SplitPDFWithSpecJSON(src, []byte(`{"pages":[1]}`))
	if err != nil {
		t.Fatalf("SplitPDFWithSpecJSON pages array failed: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	assertPDF(t, parts[0])

	strParts, err := gopdflib.SplitPDFWithSpecJSON(src, []byte(`{"pages":"1"}`))
	if err != nil {
		t.Fatalf("SplitPDFWithSpecJSON pages string failed: %v", err)
	}
	if len(strParts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(strParts))
	}
}

func TestSplitPDFWithSpecJSONMaxPerFile(t *testing.T) {
	raw := readWasmFixture(t, "..", "..", "sampledata", "financialreport", "financial_report.json")
	gen, err := gopdflib.GeneratePDFFromJSON(raw)
	if err != nil {
		t.Fatalf("GeneratePDFFromJSON failed: %v", err)
	}
	// The fixture renders 2 pages; one part per file exercises the
	// multi-file return path (JS converts each part to a Uint8Array).
	parts, err := gopdflib.SplitPDFWithSpecJSON(gen, []byte(`{"max_per_file":1}`))
	if err != nil {
		t.Fatalf("SplitPDFWithSpecJSON max_per_file failed: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}
	for i, p := range parts {
		if !bytes.HasPrefix(p, []byte("%PDF-")) {
			t.Errorf("part %d missing PDF header", i)
		}
	}
}

func TestSplitSpecJSONShapes(t *testing.T) {
	spec, err := gopdflib.ParseSplitSpecJSON([]byte(`{"pages":[1,3,5],"maxPerFile":2}`))
	if err != nil {
		t.Fatalf("array shape failed: %v", err)
	}
	if len(spec.Pages) != 3 || spec.MaxPerFile != 2 {
		t.Errorf("unexpected spec: %+v", spec)
	}

	spec, err = gopdflib.ParseSplitSpecJSON([]byte(`{"pages":"1-3,5","max_per_file":2}`))
	if err != nil {
		t.Fatalf("string shape failed: %v", err)
	}
	want := []int{1, 2, 3, 5}
	if len(spec.Pages) != len(want) {
		t.Fatalf("unexpected pages: %v", spec.Pages)
	}
	for i := range want {
		if spec.Pages[i] != want[i] {
			t.Fatalf("unexpected pages: %v", spec.Pages)
		}
	}

	spec, err = gopdflib.ParseSplitSpecJSON(nil)
	if err != nil {
		t.Fatalf("empty shape failed: %v", err)
	}
	if len(spec.Pages) != 0 || spec.MaxPerFile != 0 {
		t.Errorf("empty spec should select all pages in one file: %+v", spec)
	}

	for _, doc := range []string{
		`{"pages":"bogus"}`,
		`{"pages":[0]}`,
		`{"pages":[-2]}`,
		`{"maxPerFile":-1}`,
		`{"ranges":[[3,1]]}`,
		`{not json`,
	} {
		if _, err := gopdflib.ParseSplitSpecJSON([]byte(doc)); !errors.Is(err, gopdflib.ErrInvalidInput) {
			t.Errorf("%s: expected ErrInvalidInput, got %v", doc, err)
		}
	}
}
