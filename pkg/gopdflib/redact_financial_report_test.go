package gopdflib_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib"
)

func TestRedactFin(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "sampledata", "financialreport", "financial_report.pdf")
	pdfBytes, err := os.ReadFile(pdfPath) //nolint:gosec // test data
	if err != nil {
		t.Fatalf("failed to read sample PDF %s: %v", pdfPath, err)
	}

	opts := gopdflib.ApplyRedactionOptions{
		Mode: "secure_required",
		TextSearch: []gopdflib.RedactionTextQuery{
			{Text: "SECTION"},
			{Text: "Total"},
		},
	}

	out, report, err := gopdflib.ApplyRedactionsReport(pdfBytes, opts)
	if err != nil {
		t.Fatalf("ApplyRedactionsReport failed: %v", err)
	}
	t.Logf("redaction report: generated=%d matchedText=%d applied=%d warnings=%v capabilities=%v", report.GeneratedRects, report.MatchedTextCount, report.AppliedRectangles, report.Warnings, report.Capabilities)
	if len(out) == 0 {
		t.Fatal("redaction output is empty")
	}

	// Store output at repository root for easier inspection by developer.
	outputPath := filepath.Join("..", "..", "sampledata", "financialreport", "financial_report_redacted_gopdflib_test_output.pdf")
	if err := os.WriteFile(outputPath, out, 0o600); err != nil {
		t.Fatalf("failed to write redacted output PDF: %v", err)
	}
	t.Logf("redacted test output written to: %s", outputPath)

	if report.GeneratedRects == 0 {
		t.Fatalf("expected at least one generated redaction rectangle; report=%+v", report)
	}

	if bytes.Equal(out, pdfBytes) {
		t.Fatal("expected redacted output to differ from input when matches were generated")
	}
}

func TestRedactFinP2(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "sampledata", "financialreport", "financial_report.pdf")
	pdfBytes, err := os.ReadFile(pdfPath) //nolint:gosec // test data
	if err != nil {
		t.Fatalf("failed to read sample PDF %s: %v", pdfPath, err)
	}

	rects, err := gopdflib.FindTextOccurrences(pdfBytes, "SECTION C")
	if err != nil {
		t.Fatalf("FindTextOccurrences failed: %v", err)
	}
	if len(rects) == 0 {
		t.Fatal("expected matches for 'SECTION C' in sample PDF")
	}

	foundPage2 := false
	for _, r := range rects {
		if r.PageNum == 2 {
			foundPage2 = true
			break
		}
	}
	if !foundPage2 {
		t.Fatalf("expected at least one page 2 match for 'SECTION C', got rects=%+v", rects)
	}
}
