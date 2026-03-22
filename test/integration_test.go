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

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
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
	// Cleanup handled by individual tests or OS if needed
}

// Helper to compare file sizes
func (s *IntegrationSuite) compareFileSizes(generatedPath, expectedPath string) {
	genInfo, err := os.Stat(generatedPath)
	s.NoError(err, "Failed to stat generated file: "+generatedPath)

	expInfo, err := os.Stat(expectedPath)
	s.NoError(err, "Failed to stat expected file: "+expectedPath)

	s.Equal(expInfo.Size(), genInfo.Size(), "File sizes do not match for %s and %s", generatedPath, expectedPath)
}

// Helper to compare file sizes with a tolerance (for PDFs with timestamps/signatures)
func (s *IntegrationSuite) compareSizes(generatedPath, expectedPath string, toleranceBytes int64) {
	genInfo, err := os.Stat(generatedPath)
	s.NoError(err, "Failed to stat generated file: "+generatedPath)

	expInfo, err := os.Stat(expectedPath)
	s.NoError(err, "Failed to stat expected file: "+expectedPath)

	diff := genInfo.Size() - expInfo.Size()
	if diff < 0 {
		diff = -diff
	}
	s.LessOrEqual(diff, toleranceBytes, "File size difference %d exceeds tolerance %d for %s and %s", diff, toleranceBytes, generatedPath, expectedPath)
}

// TestGenerateTemplatePDF tests /api/v1/generate/template-pdf
func (s *IntegrationSuite) TestGenerateTemplatePDF() {
	t := s.T()
	// Error case: malformed JSON body should be rejected
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBufferString("{invalid json}"))
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("malformed JSON should not return 200, got %d", errResp.StatusCode)
	}

	// 1. Input JSON from sampledata/editor/financial_report.json
	jsonPath := filepath.Join("..", "sampledata", "editor", "financial_digitalsignature.json")
	jsonData, err := os.ReadFile(jsonPath) //nolint:gosec
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
	err = os.WriteFile(tempPath, body, 0o600)
	s.NoError(err, "Failed to write temp_editor.pdf")

	// 4. Check size against generated.pdf (use tolerance due to digital signature timestamps)
	expectedPath := filepath.Join("..", "sampledata", "editor", "generated.pdf")
	s.compareSizes(tempPath, expectedPath, 1024)
}

// TestGenerateEncryptedTemplatePDF tests encrypted template generation using the
// sample payload that enables password protection.
func (s *IntegrationSuite) TestGenerateEncryptedTemplatePDF() {
	jsonPath := filepath.Join("..", "sampledata", "editor", "financial_encrypted.json")
	jsonData, err := os.ReadFile(jsonPath) //nolint:gosec
	s.NoError(err, "Failed to read encrypted sample JSON file")

	resp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.True(bytes.HasPrefix(body, []byte("%PDF-")), "generated output should be a PDF")
	s.True(bytes.Contains(body, []byte("/Encrypt")), "generated PDF should contain an encryption dictionary")
	s.True(bytes.Contains(body, []byte("/CFM /AESV2")), "generated PDF should use AESV2 crypt filter")
	s.False(bytes.Contains(body, []byte("/OutputIntents [")), "encrypted PDF generation should disable PDF/A output intents")
}

// TestMergePDFs tests /api/v1/merge
func (s *IntegrationSuite) TestMergePDFs() {
	t := s.T()
	// Error case: merging with no files should be rejected
	emptyBody := &bytes.Buffer{}
	emptyWriter := multipart.NewWriter(emptyBody)
	require.NoError(t, emptyWriter.Close())
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/merge", emptyWriter.FormDataContentType(), emptyBody)
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("merge with no files should not return 200, got %d", errResp.StatusCode)
	}

	// 1. Inputs: em-16.pdf, em-19.pdf, em-51.pdf
	files := []string{"em-16.pdf", "em-19.pdf", "em-51.pdf"}
	baseDir := filepath.Join("..", "sampledata", "merge")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, fname := range files {
		fpath := filepath.Join(baseDir, fname)
		data, err := os.ReadFile(fpath) //nolint:gosec
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
	err = os.WriteFile(tempPath, respBody, 0o600)
	s.NoError(err, "Failed to write temp_merge.pdf")

	// 4. Check size against generated.pdf
	expectedPath := filepath.Join(baseDir, "generated.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestFillPDF tests /api/v1/fill
func (s *IntegrationSuite) TestFillPDF() {
	t := s.T()
	// Error case: fill with xfdf but no pdf field should be rejected
	errBody := &bytes.Buffer{}
	errWriter := multipart.NewWriter(errBody)
	xfdfErrPart, err := errWriter.CreateFormFile("xfdf", "data.xfdf")
	require.NoError(t, err)
	_, err = xfdfErrPart.Write([]byte(`<?xml version="1.0"?><xfdf/>`))
	require.NoError(t, err)
	require.NoError(t, errWriter.Close())
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/fill", errWriter.FormDataContentType(), errBody)
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("fill without pdf should not return 200, got %d", errResp.StatusCode)
	}

	baseDir := filepath.Join("..", "sampledata", "filler")

	// 1. Inputs
	pdfPath := filepath.Join(baseDir, "us_hospital_encounter_acroform.pdf")
	xfdfPath := filepath.Join(baseDir, "us_hospital_encounter_data.xfdf")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add PDF
	pdfData, err := os.ReadFile(pdfPath) //nolint:gosec
	if err != nil {
		s.T().Skip("Skipping TestFillPDF: sample PDF not found")
		return
	}
	pdfPart, err := writer.CreateFormFile("pdf", "form.pdf")
	s.NoError(err)
	_, err = pdfPart.Write(pdfData)
	s.NoError(err)

	// Add XFDF
	xfdfData, err := os.ReadFile(xfdfPath) //nolint:gosec
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
	err = os.WriteFile(tempPath, respBody, 0o600)
	s.NoError(err, "Failed to write temp_filler.pdf")

	// 4. Check size against generated.pdf
	expectedPath := filepath.Join(baseDir, "generated.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestHtmlToPDF tests /api/v1/htmltopdf
func (s *IntegrationSuite) TestHtmlToPDF() {
	t := s.T()
	// Error case: empty URL body should be rejected
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/htmltopdf", "application/json", bytes.NewBufferString("{}"))
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("htmltopdf with empty URL should not return 200, got %d", errResp.StatusCode)
	}

	// 1. Input URL
	req := models.HTMLToPDFRequest{
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
	s.NoError(os.MkdirAll(baseDir, 0o750))

	tempPath := filepath.Join(baseDir, "temp_htmltopdf.pdf")
	err = os.WriteFile(tempPath, body, 0o600)
	s.NoError(err, "Failed to write temp_htmltopdf.pdf")

	// 4. Check that the file is non-zero size
	info, err := os.Stat(tempPath)
	s.NoError(err)
	s.Greater(info.Size(), int64(0), "Generated PDF should have non-zero size")
}

// TestHtmlToImage tests /api/v1/htmltoimage
func (s *IntegrationSuite) TestHtmlToImage() {
	t := s.T()
	// Error case: empty URL body should be rejected
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/htmltoimage", "application/json", bytes.NewBufferString("{}"))
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("htmltoimage with empty URL should not return 200, got %d", errResp.StatusCode)
	}

	// 1. Input URL
	req := models.HTMLToImageRequest{
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
	s.NoError(os.MkdirAll(baseDir, 0o750))

	tempPath := filepath.Join(baseDir, "temp_htmltoimage.png")
	err = os.WriteFile(tempPath, body, 0o600)
	s.NoError(err, "Failed to write temp_htmltoimage.png")

	// 4. Check that the file is non-zero size
	info, err := os.Stat(tempPath)
	s.NoError(err)
	s.Greater(info.Size(), int64(0), "Generated image should have non-zero size")
}

// TestSplitPDF tests /api/v1/split with single page
func (s *IntegrationSuite) TestSplitPDF() {
	t := s.T()
	// Error case: split with no pdf file should be rejected
	errBody := &bytes.Buffer{}
	errWriter := multipart.NewWriter(errBody)
	require.NoError(t, errWriter.WriteField("pages", "1"))
	require.NoError(t, errWriter.Close())
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/split", errWriter.FormDataContentType(), errBody)
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("split without pdf should not return 200, got %d", errResp.StatusCode)
	}

	baseDir := filepath.Join("..", "sampledata", "split")

	// 1. Input PDF
	pdfPath := filepath.Join(baseDir, "em.pdf")
	pdfData, err := os.ReadFile(pdfPath) //nolint:gosec
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
	err = os.WriteFile(tempPath, respBody, 0o600)
	s.NoError(err, "Failed to write temp_split.pdf")

	// 4. Check size against split.pdf
	expectedPath := filepath.Join(baseDir, "split.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestSplitPDFRange tests /api/v1/split with page range
func (s *IntegrationSuite) TestSplitPDFRange() {
	t := s.T()
	// Error case: split with invalid page range should be rejected
	errBody := &bytes.Buffer{}
	errWriter := multipart.NewWriter(errBody)
	errPart, err := errWriter.CreateFormFile("pdf", "fake.pdf")
	require.NoError(t, err)
	_, err = errPart.Write([]byte("not a pdf"))
	require.NoError(t, err)
	require.NoError(t, errWriter.WriteField("pages", "abc-xyz"))
	require.NoError(t, errWriter.Close())
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/split", errWriter.FormDataContentType(), errBody)
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("split with invalid page range should not return 200, got %d", errResp.StatusCode)
	}

	baseDir := filepath.Join("..", "sampledata", "split")

	// 1. Input PDF
	pdfPath := filepath.Join(baseDir, "em.pdf")
	pdfData, err := os.ReadFile(pdfPath) //nolint:gosec
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
	err = os.WriteFile(tempPath, respBody, 0o600)
	s.NoError(err, "Failed to write temp_split_range.pdf")

	// 4. Check size against split_range.pdf
	expectedPath := filepath.Join(baseDir, "split_range.pdf")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestSplitPDFMaxPerFile tests /api/v1/split with max per file
func (s *IntegrationSuite) TestSplitPDFMaxPerFile() {
	t := s.T()
	// Error case: non-numeric max_per_file should be rejected
	errBody := &bytes.Buffer{}
	errWriter := multipart.NewWriter(errBody)
	errPart, err := errWriter.CreateFormFile("pdf", "fake.pdf")
	require.NoError(t, err)
	_, err = errPart.Write([]byte("not a pdf"))
	require.NoError(t, err)
	require.NoError(t, errWriter.WriteField("pages", "1-3"))
	require.NoError(t, errWriter.WriteField("max_per_file", "abc"))
	require.NoError(t, errWriter.Close())
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/split", errWriter.FormDataContentType(), errBody)
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("split with non-numeric max_per_file should not return 200, got %d", errResp.StatusCode)
	}

	baseDir := filepath.Join("..", "sampledata", "split")

	// 1. Input PDF
	pdfPath := filepath.Join(baseDir, "em.pdf")
	pdfData, err := os.ReadFile(pdfPath) //nolint:gosec
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
	err = os.WriteFile(tempPath, respBody, 0o600)
	s.NoError(err, "Failed to write temp_maxperfile.zip")

	// 4. Check size against maxperfile.zip
	expectedPath := filepath.Join(baseDir, "maxperfile.zip")
	s.compareFileSizes(tempPath, expectedPath)
}

// TestTypstMathShow tests /api/v1/generate/template-pdf with typst_math_showcase.json
func (s *IntegrationSuite) TestTypstMathShow() {
	t := s.T()
	// Error case: malformed JSON body should be rejected
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBufferString("{bad}"))
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("malformed JSON for typst should not return 200, got %d", errResp.StatusCode)
	}

	baseDir := filepath.Join("..", "sampledata", "typstsyntax")
	jsonPath := filepath.Join(baseDir, "typst_math_showcase.json")
	jsonData, err := os.ReadFile(jsonPath) //nolint:gosec
	s.NoError(err, "Failed to read typst_math_showcase.json")

	resp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBuffer(jsonData))
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Greater(len(body), 0, "Generated PDF should not be empty")

	outPath := filepath.Join(baseDir, "typst_math_showcase.pdf")
	err = os.WriteFile(outPath, body, 0o600)
	s.NoError(err, "Failed to write typst_math_showcase.pdf")

	info, err := os.Stat(outPath)
	s.NoError(err)
	s.Greater(info.Size(), int64(0), "typst_math_showcase.pdf should have non-zero size")
}

// TestTypstSample tests /api/v1/generate/template-pdf with typst_sample.json
func (s *IntegrationSuite) TestTypstSample() {
	t := s.T()
	// Error case: empty template body should be rejected
	errResp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBufferString("{}"))
	require.NoError(t, err)
	require.NoError(t, errResp.Body.Close())
	if errResp.StatusCode == http.StatusOK {
		t.Fatalf("empty template body should not return 200, got %d", errResp.StatusCode)
	}

	mathFontPath, ok := resolveMathPath()
	if !ok {
		s.T().Skip("no unicode math-capable font found (looked for DejaVu/Noto Math). Install fonts-dejavu-core or fonts-noto-math")
	}

	baseDir := filepath.Join("..", "sampledata", "typstsyntax")
	jsonPath := filepath.Join(baseDir, "typst_sample.json")
	jsonData, err := os.ReadFile(jsonPath) //nolint:gosec
	s.NoError(err, "Failed to read typst_sample.json")

	// Unmarshal, inject customFonts config with resolved font path, then re-marshal
	var template models.PDFTemplate
	err = json.Unmarshal(jsonData, &template)
	s.NoError(err, "Failed to unmarshal typst_sample.json")

	template.Config.CustomFonts = []models.CustomFontConfig{{
		Name:     "MathUnicode",
		FilePath: mathFontPath,
	}}

	modifiedJSON, err := json.Marshal(template)
	s.NoError(err, "Failed to marshal modified template")

	resp, err := s.client.Post(s.ts.URL+"/api/v1/generate/template-pdf", "application/json", bytes.NewBuffer(modifiedJSON))
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Greater(len(body), 0, "Generated PDF should not be empty")

	outPath := filepath.Join(baseDir, "typst_sample.pdf")
	err = os.WriteFile(outPath, body, 0o600)
	s.NoError(err, "Failed to write typst_sample.pdf")

	info, err := os.Stat(outPath)
	s.NoError(err)
	s.Greater(info.Size(), int64(0), "typst_sample.pdf should have non-zero size")
}

// resolveMathPath finds a unicode-capable math font on the system
func resolveMathPath() (string, bool) {
	candidates := []string{
		"/usr/share/fonts/truetype/noto/NotoSansMath-Regular.ttf",
		"/usr/share/fonts/opentype/noto/NotoSansMath-Regular.otf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

// Run the suite
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests require a running HTTP server setup")
	}
	s := new(IntegrationSuite)
	suite.Run(t, s)
}
