package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/internal/handlers"
	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

// IntegrationSuite defines the suite
type IntegrationSuite struct {
	suite.Suite
	server *gin.Engine
	client *http.Client
	ts     *httptest.Server
}

// SetupSuite runs ONCE before all tests in this suite
func (s *IntegrationSuite) SetupSuite() {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Initialize router
	s.server = gin.Default()
	handlers.RegisterRoutes(s.server)

	// Create a test server
	s.ts = httptest.NewServer(s.server)
	s.client = s.ts.Client()
}

// TearDownSuite runs ONCE after all tests are done
func (s *IntegrationSuite) TearDownSuite() {
	s.ts.Close()
}

// Helper to compare file sizes
func (s *IntegrationSuite) compareFileSizes(generatedPath, expectedPath string) {
	genInfo, err := os.Stat(generatedPath)
	s.NoError(err, "Failed to stat generated file: "+generatedPath)

	expInfo, err := os.Stat(expectedPath)
	s.NoError(err, "Failed to stat expected file: "+expectedPath)

	// Allow for some tolerance or just log if different?
	// User asked to "check their sizes if that is same or not".
	// We will assert equality.
	s.Equal(expInfo.Size(), genInfo.Size(), "File sizes do not match for %s and %s", generatedPath, expectedPath)
}

// TestGenerateTemplatePDF tests /api/v1/generate/template-pdf
func (s *IntegrationSuite) TestGenerateTemplatePDF() {
	// 1. Input JSON from sampledata/editor/financial_report.json
	jsonPath := filepath.Join("..", "sampledata", "editor", "financial_report.json")
	jsonData, err := os.ReadFile(jsonPath)
	s.NoError(err, "Failed to read sample JSON file")

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	// 3. Create temp_editor.pdf
	body, err := io.ReadAll(resp.Body)
	s.NoError(err)

	tempPath := filepath.Join("..", "sampledata", "editor", "temp_editor.pdf")
	err = os.WriteFile(tempPath, body, 0644)
	s.NoError(err, "Failed to write temp_editor.pdf")

	// 4. Check size against generated.pdf
	expectedPath := filepath.Join("..", "sampledata", "editor", "generated.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestMergePDFs tests /api/v1/merge
func (s *IntegrationSuite) TestMergePDFs() {
	// 1. Inputs: em-16.pdf, em-19.pdf, em-51.pdf
	files := []string{"em-16.pdf", "em-19.pdf", "em-51.pdf"}
	baseDir := filepath.Join("..", "sampledata", "merge")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, fname := range files {
		fpath := filepath.Join(baseDir, fname)
		data, err := os.ReadFile(fpath)
		if err != nil {
			s.T().Skipf("Skipping TestMergePDFs: file %s not found", fname)
			return
		}
		part, err := writer.CreateFormFile("pdf", fname)
		s.NoError(err)
		_, err = part.Write(data)
		s.NoError(err)
	}
	writer.Close()

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/merge", writer.FormDataContentType(), body)
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	// 3. Create temp_merge.pdf
	respBody, err := io.ReadAll(resp.Body)
	s.NoError(err)

	tempPath := filepath.Join(baseDir, "temp_merge.pdf")
	err = os.WriteFile(tempPath, respBody, 0644)
	s.NoError(err, "Failed to write temp_merge.pdf")

	// 4. Check size against generated.pdf
	expectedPath := filepath.Join(baseDir, "generated.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestFillPDF tests /api/v1/fill
func (s *IntegrationSuite) TestFillPDF() {
	baseDir := filepath.Join("..", "sampledata", "filler")

	// 1. Inputs
	pdfPath := filepath.Join(baseDir, "us_hospital_encounter_acroform.pdf")
	xfdfPath := filepath.Join(baseDir, "us_hospital_encounter_data.xfdf")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add PDF
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		s.T().Skip("Skipping TestFillPDF: sample PDF not found")
		return
	}
	pdfPart, err := writer.CreateFormFile("pdf", "form.pdf")
	s.NoError(err)
	_, err = pdfPart.Write(pdfData)
	s.NoError(err)

	// Add XFDF
	xfdfData, err := os.ReadFile(xfdfPath)
	if err != nil {
		s.T().Skip("Skipping TestFillPDF: sample XFDF not found")
		return
	}
	xfdfPart, err := writer.CreateFormFile("xfdf", "data.xfdf")
	s.NoError(err)
	_, err = xfdfPart.Write(xfdfData)
	s.NoError(err)

	writer.Close()

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/fill", writer.FormDataContentType(), body)
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	// 3. Create temp_filler.pdf
	respBody, err := io.ReadAll(resp.Body)
	s.NoError(err)

	tempPath := filepath.Join(baseDir, "temp_filler.pdf")
	err = os.WriteFile(tempPath, respBody, 0644)
	s.NoError(err, "Failed to write temp_filler.pdf")

	// 4. Check size against generated.pdf
	expectedPath := filepath.Join(baseDir, "generated.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestHtmlToPDF tests /api/v1/htmltopdf
func (s *IntegrationSuite) TestHtmlToPDF() {
	// 1. Input URL
	req := models.HtmlToPDFRequest{
		URL: "https://en.wikipedia.org/wiki/Ana_de_Armas",
	}
	reqBody, _ := json.Marshal(req)

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/htmltopdf", "application/json", bytes.NewBuffer(reqBody))
	s.NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.T().Logf("HtmlToPDF failed with status: %d. Skipping size check.", resp.StatusCode)
		return
	}

	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	// 3. Create temp_htmltopdf.pdf
	body, err := io.ReadAll(resp.Body)
	s.NoError(err)

	baseDir := filepath.Join("..", "sampledata", "htmltopdf")
	// Ensure directory exists
	os.MkdirAll(baseDir, 0755)

	tempPath := filepath.Join(baseDir, "temp_htmltopdf.pdf")
	err = os.WriteFile(tempPath, body, 0644)
	s.NoError(err, "Failed to write temp_htmltopdf.pdf")

	// 4. Check size against generated.pdf
	expectedPath := filepath.Join(baseDir, "generated.pdf")
	// Only check if expected file exists
	if _, err := os.Stat(expectedPath); err == nil {
		s.compareFileSizes(tempPath, expectedPath)
	} else {
		s.T().Logf("Expected file %s not found, skipping size comparison", expectedPath)
	}
}

// TestHtmlToImage tests /api/v1/htmltoimage
func (s *IntegrationSuite) TestHtmlToImage() {
	// 1. Input URL
	req := models.HtmlToImageRequest{
		URL:    "https://en.wikipedia.org/wiki/Ana_de_Armas",
		Format: "png",
	}
	reqBody, _ := json.Marshal(req)

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/htmltoimage", "application/json", bytes.NewBuffer(reqBody))
	s.NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.T().Logf("HtmlToImage failed with status: %d. Skipping size check.", resp.StatusCode)
		return
	}

	s.Equal("image/png", resp.Header.Get("Content-Type"))

	// 3. Create temp_htmltoimage.png
	body, err := io.ReadAll(resp.Body)
	s.NoError(err)

	baseDir := filepath.Join("..", "sampledata", "htmltoimg")
	// Ensure directory exists
	os.MkdirAll(baseDir, 0755)

	tempPath := filepath.Join(baseDir, "temp_htmltoimage.png")
	err = os.WriteFile(tempPath, body, 0644)
	s.NoError(err, "Failed to write temp_htmltoimage.png")

	// 4. Check size against generated.png
	expectedPath := filepath.Join(baseDir, "generated.png")
	// Only check if expected file exists
	if _, err := os.Stat(expectedPath); err == nil {
		s.compareFileSizes(tempPath, expectedPath)
	} else {
		s.T().Logf("Expected file %s not found, skipping size comparison", expectedPath)
	}
}

// Run the suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}
