package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// TestGetFonts tests GET /api/v1/fonts
func (s *IntegrationSuite) TestGetFonts() {
	resp, err := s.client.Get(s.ts.URL + "/api/v1/fonts")
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	var payload map[string]any
	s.NoError(json.NewDecoder(resp.Body).Decode(&payload))
	s.Contains(payload, "fonts")
}

// TestGetTemplateData tests GET /api/v1/template-data for financial_report.json
func (s *IntegrationSuite) TestGetTemplateData() {
	jsonPath := samplePath(s.T(), "financialreport", "financial_report.json")
	dir := filepath.Dir(jsonPath)

	oldRoot := os.Getenv("GOPDFSUIT_ROOT")
	s.NoError(os.Setenv("GOPDFSUIT_ROOT", dir))
	defer func() { _ = os.Setenv("GOPDFSUIT_ROOT", oldRoot) }()

	resp, err := s.client.Get(s.ts.URL + "/api/v1/template-data?file=financial_report.json")
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Greater(len(body), 0)

	var tmpl map[string]any
	s.NoError(json.Unmarshal(body, &tmpl))
	s.Contains(tmpl, "config")
}

// TestUploadFontInvalidExtension tests POST /api/v1/fonts rejects non-font files.
func (s *IntegrationSuite) TestUploadFontInvalidExtension() {
	body := &bytes.Buffer{}
	// Minimal invalid upload without real multipart font - expect 400
	resp, err := s.client.Post(s.ts.URL+"/api/v1/fonts", "application/octet-stream", body)
	s.NoError(err)
	defer func() { _ = resp.Body.Close() }()
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}
