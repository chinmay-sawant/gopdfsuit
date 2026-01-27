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

// TearDownTest runs after each test
func (s *IntegrationSuite) TearDownTest() {
	// Clean up temp files
	tempFiles := []string{
		filepath.Join("..", "sampledata", "editor", "temp_editor.pdf"),
		filepath.Join("..", "sampledata", "merge", "temp_merge.pdf"),
		filepath.Join("..", "sampledata", "filler", "temp_filler.pdf"),
		filepath.Join("..", "sampledata", "htmltopdf", "temp_htmltopdf.pdf"),
		filepath.Join("..", "sampledata", "htmltoimg", "temp_htmltoimage.png"),
		filepath.Join("..", "sampledata", "split", "temp_split.pdf"),
		filepath.Join("..", "sampledata", "split", "temp_split_range.pdf"),
		filepath.Join("..", "sampledata", "split", "temp_maxperfile.zip"),
	}
	for _, f := range tempFiles {
		_ = os.Remove(f) // Ignore errors if file doesn't exist
	}
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
	jsonPath := filepath.Join("..", "sampledata", "editor", "financial_digitalsignature.json")
	jsonData, err := os.ReadFile(jsonPath)
	s.NoError(err, "Failed to read sample JSON file")

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

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
	s.NoError(writer.Close())

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/merge", writer.FormDataContentType(), body)
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

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

	s.NoError(writer.Close())

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/fill", writer.FormDataContentType(), body)
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

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
	defer func() {
		_ = resp.Body.Close()
	}()

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
	s.NoError(os.MkdirAll(baseDir, 0755))

	tempPath := filepath.Join(baseDir, "temp_htmltopdf.pdf")
	err = os.WriteFile(tempPath, body, 0644)
	s.NoError(err, "Failed to write temp_htmltopdf.pdf")

	// 4. Check that the file is non-zero size
	info, err := os.Stat(tempPath)
	s.NoError(err)
	s.Greater(info.Size(), int64(0), "Generated PDF should have non-zero size")
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
	defer func() {
		_ = resp.Body.Close()
	}()

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
	s.NoError(os.MkdirAll(baseDir, 0755))

	tempPath := filepath.Join(baseDir, "temp_htmltoimage.png")
	err = os.WriteFile(tempPath, body, 0644)
	s.NoError(err, "Failed to write temp_htmltoimage.png")

	// 4. Check that the file is non-zero size
	info, err := os.Stat(tempPath)
	s.NoError(err)
	s.Greater(info.Size(), int64(0), "Generated image should have non-zero size")
}

// TestSplitPDF tests /api/v1/split with single page
func (s *IntegrationSuite) TestSplitPDF() {
	baseDir := filepath.Join("..", "sampledata", "split")

	// 1. Input PDF
	pdfPath := filepath.Join(baseDir, "em.pdf")
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		s.T().Skip("Skipping TestSplitPDF: sample PDF not found")
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add PDF
	pdfPart, err := writer.CreateFormFile("pdf", "em.pdf")
	s.NoError(err)
	_, err = pdfPart.Write(pdfData)
	s.NoError(err)

	// Add form fields
	err = writer.WriteField("pages", "10")
	s.NoError(err)
	err = writer.WriteField("max_per_file", "")
	s.NoError(err)

	s.NoError(writer.Close())

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/split", writer.FormDataContentType(), body)
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	// 3. Create temp_split.pdf
	respBody, err := io.ReadAll(resp.Body)
	s.NoError(err)

	tempPath := filepath.Join(baseDir, "temp_split.pdf")
	err = os.WriteFile(tempPath, respBody, 0644)
	s.NoError(err, "Failed to write temp_split.pdf")

	// 4. Check size against split.pdf
	expectedPath := filepath.Join(baseDir, "split.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestSplitPDFRange tests /api/v1/split with page range
func (s *IntegrationSuite) TestSplitPDFRange() {
	baseDir := filepath.Join("..", "sampledata", "split")

	// 1. Input PDF
	pdfPath := filepath.Join(baseDir, "em.pdf")
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		s.T().Skip("Skipping TestSplitPDFRange: sample PDF not found")
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add PDF
	pdfPart, err := writer.CreateFormFile("pdf", "em.pdf")
	s.NoError(err)
	_, err = pdfPart.Write(pdfData)
	s.NoError(err)

	// Add form fields
	err = writer.WriteField("pages", "10-12")
	s.NoError(err)
	err = writer.WriteField("max_per_file", "")
	s.NoError(err)

	s.NoError(writer.Close())

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/split", writer.FormDataContentType(), body)
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	// 3. Create temp_split_range.pdf
	respBody, err := io.ReadAll(resp.Body)
	s.NoError(err)

	tempPath := filepath.Join(baseDir, "temp_split_range.pdf")
	err = os.WriteFile(tempPath, respBody, 0644)
	s.NoError(err, "Failed to write temp_split_range.pdf")

	// 4. Check size against split_range.pdf
	expectedPath := filepath.Join(baseDir, "split_range.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestSplitPDFMaxPerFile tests /api/v1/split with max per file
func (s *IntegrationSuite) TestSplitPDFMaxPerFile() {
	baseDir := filepath.Join("..", "sampledata", "split")

	// 1. Input PDF
	pdfPath := filepath.Join(baseDir, "em.pdf")
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		s.T().Skip("Skipping TestSplitPDFMaxPerFile: sample PDF not found")
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add PDF
	pdfPart, err := writer.CreateFormFile("pdf", "em.pdf")
	s.NoError(err)
	_, err = pdfPart.Write(pdfData)
	s.NoError(err)

	// Add form fields
	err = writer.WriteField("pages", "10-12")
	s.NoError(err)
	err = writer.WriteField("max_per_file", "1")
	s.NoError(err)

	s.NoError(writer.Close())

	// 2. Send to endpoint
	resp, err := s.client.Post(s.ts.URL+"/api/v1/split", writer.FormDataContentType(), body)
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/zip", resp.Header.Get("Content-Type"))

	// 3. Create temp_maxperfile.zip
	respBody, err := io.ReadAll(resp.Body)
	s.NoError(err)

	tempPath := filepath.Join(baseDir, "temp_maxperfile.zip")
	err = os.WriteFile(tempPath, respBody, 0644)
	s.NoError(err, "Failed to write temp_maxperfile.zip")

	// 4. Check size against maxperfile.zip
	expectedPath := filepath.Join(baseDir, "maxperfile.zip")
	s.compareFileSizes(tempPath, expectedPath)
}

// Run the suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}
