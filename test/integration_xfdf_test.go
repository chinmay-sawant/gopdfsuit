package tests

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// TestFillPDFCompressed tests XFDF fill on a zlib-compressed AcroForm PDF.
func (s *IntegrationSuite) TestFillPDFCompressed() {
	baseDir := samplePath(s.T(), "filler", "compressed")
	pdfPath := filepath.Join(baseDir, "medical_form.pdf")
	xfdfPath := filepath.Join(baseDir, "medical_data.xfdf")

	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		s.T().Skip("Skipping TestFillPDFCompressed: medical_form.pdf not found")
		return
	}
	xfdfData, err := os.ReadFile(xfdfPath)
	if err != nil {
		s.T().Skip("Skipping TestFillPDFCompressed: medical_data.xfdf not found")
		return
	}

	body, contentType := writeMultipartPDFXfdf(s.T(), pdfData, xfdfData)
	resp, err := s.client.Post(s.ts.URL+"/api/v1/fill", contentType, body)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	respBody, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Greater(len(respBody), 0)
	s.Equal("%PDF-", string(respBody[:5]))

	outPath := filepath.Join(baseDir, "temp_filler_compressed.pdf")
	err = os.WriteFile(outPath, respBody, 0644)
	s.NoError(err)

	expectedPath := filepath.Join(baseDir, "generated.pdf")
	if _, err := os.Stat(expectedPath); err == nil {
		s.compareFileSizesWithTolerance(outPath, expectedPath, 700)
	}

	assertFilledXFDFFields(s.T(), respBody, xfdfData)
}

// TestFillPDFHospitalEncounter is an explicit alias-style test for the standard hospital XFDF sample.
func (s *IntegrationSuite) TestFillPDFHospitalEncounter() {
	s.TestFillPDF()
}
