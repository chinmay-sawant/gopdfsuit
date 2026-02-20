package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFinancialReportTextRedaction(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "sampledata", "financialreport", "financial_report.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("failed to read sample PDF %s: %v", pdfPath, err)
	}

	opts := ApplyRedactionOptions{
		Mode: "visual_allowed",
		TextSearch: []RedactionTextQuery{
			{Text: "A"},
		},
	}

	out, report, err := ApplyRedactionsAdvancedWithReport(pdfBytes, opts)
	if err != nil {
		t.Fatalf("ApplyRedactionsAdvancedWithReport failed: %v", err)
	}
	t.Logf("redaction report: generated=%d matchedText=%d applied=%d warnings=%v capabilities=%v", report.GeneratedRects, report.MatchedTextCount, report.AppliedRectangles, report.Warnings, report.Capabilities)
	if len(out) == 0 {
		t.Fatal("redaction output is empty")
	}

	// store output at repository root for easier inspection by developer
	outputPath := filepath.Join("..", "..", "financial_report_redacted_test_output.pdf")
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
