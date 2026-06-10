package tests

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// TestGenerateFinancialReportPDF generates from sampledata/financialreport/financial_report.json
// and writes financial_report.pdf in the same directory.
func (s *IntegrationSuite) TestGenerateFinancialReportPDF() {
	jsonPath := samplePath(s.T(), "financialreport", "financial_report.json")
	jsonData, err := os.ReadFile(jsonPath)
	s.NoError(err, "Failed to read financial_report.json")

	resp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Greater(len(body), 0)
	s.Equal("%PDF-", string(body[:5]))

	outPath := filepath.Join(filepath.Dir(jsonPath), "financial_report.pdf")
	err = os.WriteFile(outPath, body, 0644)
	s.NoError(err, "Failed to write financial_report.pdf")

	info, err := os.Stat(outPath)
	s.NoError(err)
	s.Greater(info.Size(), int64(1000), "financial_report.pdf should be substantial")
}