package redact

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

func TestFinancialReportTextRedaction(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "..", "sampledata", "financialreport", "compliant_financial_report.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("failed to read sample PDF %s: %v", pdfPath, err)
	}

	opts := models.ApplyRedactionOptions{
		Mode: "secure_required",
		TextSearch: []models.RedactionTextQuery{
			{Text: "SECTION"},
			{Text: "Total"},
		},
	}

	r, err := NewRedactor(pdfBytes)
	if err != nil {
		t.Fatalf("NewRedactor failed: %v", err)
	}
	out, report, err := r.ApplyRedactionsAdvancedWithReport(opts)
	if err != nil {
		t.Fatalf("ApplyRedactionsAdvancedWithReport failed: %v", err)
	}
	t.Logf("redaction report: generated=%d matchedText=%d applied=%d warnings=%v capabilities=%v", report.GeneratedRects, report.MatchedTextCount, report.AppliedRectangles, report.Warnings, report.Capabilities)
	if len(out) == 0 {
		t.Fatal("redaction output is empty")
	}

	// store output at repository root for easier inspection by developer
	outputPath := filepath.Join("..", "..", "..", "sampledata", "financialreport", "compliant_financial_report_redacted.pdf")
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

func TestFinancialReportPage2TextRedaction(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "..", "sampledata", "financialreport", "compliant_financial_report.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("failed to read sample PDF %s: %v", pdfPath, err)
	}

	r, err := NewRedactor(pdfBytes)
	if err != nil {
		t.Fatalf("NewRedactor failed: %v", err)
	}
	rects, err := r.FindTextOccurrences("SECTION C")
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

func TestFinancialReportPage2TextRedactionMultiTerms(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "..", "sampledata", "financialreport", "compliant_financial_report.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("failed to read sample PDF %s: %v", pdfPath, err)
	}

	r, err := NewRedactor(pdfBytes)
	if err != nil {
		t.Fatalf("NewRedactor failed: %v", err)
	}
	rects, err := r.FindTextOccurrencesMulti([]string{"SEC", "COM"})
	if err != nil {
		t.Fatalf("FindTextOccurrencesMulti failed: %v", err)
	}
	if len(rects) == 0 {
		t.Fatal("expected matches for SEC/COM in sample PDF")
	}

	foundPage2 := false
	for _, r := range rects {
		if r.PageNum == 2 {
			foundPage2 = true
			break
		}
	}
	if !foundPage2 {
		t.Fatalf("expected at least one page 2 match for SEC/COM, got rects=%+v", rects)
	}
}
