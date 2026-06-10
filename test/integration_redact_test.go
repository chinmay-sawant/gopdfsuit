package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (s *IntegrationSuite) loadFinancialReportPDF() []byte {
	s.T().Helper()
	pdfPath := samplePath(s.T(), "financialreport", "financial_report.pdf")
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		s.T().Skipf("financial_report.pdf not found: %v", err)
	}
	return data
}

// TestRedactPageInfo tests POST /api/v1/redact/page-info
func (s *IntegrationSuite) TestRedactPageInfo() {
	pdfData := s.loadFinancialReportPDF()
	body, contentType := writeMultipartPDF(s.T(), "pdf", "financial_report.pdf", pdfData, nil)

	resp, err := s.client.Post(s.ts.URL+"/api/v1/redact/page-info", contentType, body)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Contains(resp.Header.Get("Content-Type"), "application/json")

	var info map[string]any
	s.NoError(json.NewDecoder(resp.Body).Decode(&info))
	s.NotEmpty(info)
}

// TestRedactCapabilities tests POST /api/v1/redact/capabilities
func (s *IntegrationSuite) TestRedactCapabilities() {
	pdfData := s.loadFinancialReportPDF()
	body, contentType := writeMultipartPDF(s.T(), "pdf", "financial_report.pdf", pdfData, nil)

	resp, err := s.client.Post(s.ts.URL+"/api/v1/redact/capabilities", contentType, body)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	var payload map[string]any
	s.NoError(json.NewDecoder(resp.Body).Decode(&payload))
	s.Contains(payload, "capabilities")
}

// TestRedactTextPositions tests POST /api/v1/redact/text-positions
func (s *IntegrationSuite) TestRedactTextPositions() {
	pdfData := s.loadFinancialReportPDF()
	body, contentType := writeMultipartPDF(s.T(), "pdf", "financial_report.pdf", pdfData, map[string]string{
		"page": "1",
	})

	resp, err := s.client.Post(s.ts.URL+"/api/v1/redact/text-positions", contentType, body)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

// TestRedactSearch tests POST /api/v1/redact/search
func (s *IntegrationSuite) TestRedactSearch() {
	pdfData := s.loadFinancialReportPDF()
	body, contentType := writeMultipartPDF(s.T(), "pdf", "financial_report.pdf", pdfData, map[string]string{
		"text": "Revenue",
	})

	resp, err := s.client.Post(s.ts.URL+"/api/v1/redact/search", contentType, body)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

// TestRedactApply tests POST /api/v1/redact/apply with text search redaction.
func (s *IntegrationSuite) TestRedactApply() {
	pdfData := s.loadFinancialReportPDF()
	body, contentType := writeMultipartPDF(s.T(), "pdf", "financial_report.pdf", pdfData, map[string]string{
		"textSearch": `[{"text":"SECTION"},{"text":"Total"}]`,
		"mode":       "secure_required",
	})

	resp, err := s.client.Post(s.ts.URL+"/api/v1/redact/apply", contentType, body)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	out, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Greater(len(out), 0)

	outPath := filepath.Join(samplePath(s.T(), "financialreport"), "temp_financial_report_redacted.pdf")
	s.NoError(os.WriteFile(outPath, out, 0644))
}