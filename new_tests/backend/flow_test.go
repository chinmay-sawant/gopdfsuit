package backend_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	handlers.RegisterRoutes(r)
	return r
}

func TestFlowCases(t *testing.T) {
	router := setupRouter()

	// Helper to find the project root from test directory
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	t.Run("TC 01: Verificar PDF rellenable en /api/v1/redact/capabilities (Pass)", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Create form file
		part, err := writer.CreateFormFile("pdf", "us_patient_healthcare_form_compressed.pdf")
		if err != nil {
			t.Fatal(err)
		}

		pdfPath := filepath.Join(projectRoot, "sampledata", "acroform", "us_patient_healthcare_form_compressed.pdf")
		fileData, err := os.ReadFile(pdfPath)
		if err != nil {
			t.Skipf("File not found, skipping: %v", err)
			return
		}

		part.Write(fileData)
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/v1/redact/capabilities", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %v. Body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Ensure it responds with capabilities array
		if caps, ok := resp["capabilities"].([]interface{}); !ok || len(caps) == 0 {
			t.Errorf("Expected capabilities array, got %v", resp["capabilities"])
		}
		if fill, ok := resp["is_fillable"].(bool); !ok || !fill {
			t.Errorf("Expected is_fillable true for AcroForm sample, got %v", resp["is_fillable"])
		}
	})

	t.Run("TC 04: Rellenar formulario independiente en /api/v1/fill (Pass)", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// PDF
		partPdf, _ := writer.CreateFormFile("pdf", "form.pdf")
		pdfPath := filepath.Join(projectRoot, "sampledata", "acroform", "us_patient_healthcare_form_compressed.pdf")
		pdfData, err := os.ReadFile(pdfPath)
		if err == nil {
			partPdf.Write(pdfData)
		} else {
			// fallback mock
			partPdf.Write([]byte("%PDF-1.4\n%EOF\n"))
		}

		// XFDF
		partXfdf, _ := writer.CreateFormFile("xfdf", "data.xfdf")
		partXfdf.Write([]byte("<xfdf xmlns=\"http://ns.adobe.com/xfdf/\"><fields></fields></xfdf>"))

		writer.Close()

		req, _ := http.NewRequest("POST", "/api/v1/fill", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %v. Body: %s", w.Code, w.Body.String())
		}
		if w.Header().Get("Content-Type") != "application/pdf" {
			t.Errorf("Expected application/pdf, got %v", w.Header().Get("Content-Type"))
		}
		if len(w.Body.Bytes()) == 0 {
			t.Errorf("Expected non-empty PDF body")
		}
	})

	t.Run("TC 06: Generación de portada con payload vacío en /generate/template-pdf (Fail)", func(t *testing.T) {
		reqEmpty, _ := http.NewRequest("POST", "/api/v1/generate/template-pdf", nil)
		reqEmpty.Header.Set("Content-Type", "application/json")

		wEmpty := httptest.NewRecorder()
		router.ServeHTTP(wEmpty, reqEmpty)

		if wEmpty.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty body, got %v. Body: %s", wEmpty.Code, wEmpty.Body.String())
		}

		reqBad, _ := http.NewRequest("POST", "/api/v1/generate/template-pdf", bytes.NewBufferString("{ invalid json "))
		reqBad.Header.Set("Content-Type", "application/json")

		wBad := httptest.NewRecorder()
		router.ServeHTTP(wBad, reqBad)

		if wBad.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for bad JSON, got %v. Body: %s", wBad.Code, wBad.Body.String())
		}
	})

	t.Run("TC 08: Fusión exitosa en /api/v1/merge (Pass)", func(t *testing.T) {
		pdfPath := filepath.Join(projectRoot, "sampledata", "acroform", "us_patient_healthcare_form_compressed.pdf")
		pdfData, err := os.ReadFile(pdfPath)
		if err != nil {
			t.Skipf("sample PDF not found: %v", err)
		}

		bodyValid := &bytes.Buffer{}
		writerValid := multipart.NewWriter(bodyValid)
		p1, _ := writerValid.CreateFormFile("pdf", "valid1.pdf")
		p1.Write(pdfData)
		p2, _ := writerValid.CreateFormFile("pdf", "valid2.pdf")
		p2.Write(pdfData)
		writerValid.Close()

		reqValid, _ := http.NewRequest("POST", "/api/v1/merge", bodyValid)
		reqValid.Header.Set("Content-Type", writerValid.FormDataContentType())

		wValid := httptest.NewRecorder()
		router.ServeHTTP(wValid, reqValid)

		if wValid.Code != http.StatusOK {
			t.Errorf("Expected status 200 for valid PDFs merge, got %v. Body: %s", wValid.Code, wValid.Body.String())
		}
		if ct := wValid.Header().Get("Content-Type"); ct != "application/pdf" {
			t.Errorf("Expected application/pdf, got %q", ct)
		}
	})

	t.Run("TC 09: Validación en /api/v1/redact/page-info con bundle válido (Pass)", func(t *testing.T) {
		pdfPath := filepath.Join(projectRoot, "sampledata", "acroform", "us_patient_healthcare_form_compressed.pdf")
		pdfData, err := os.ReadFile(pdfPath)
		if err != nil {
			t.Skip("File not found")
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("pdf", "encrypted.pdf")
		part.Write(pdfData)
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/v1/redact/page-info", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %v. Body: %s", w.Code, w.Body.String())
		}

		var pageInfo models.PageInfo
		if err := json.Unmarshal(w.Body.Bytes(), &pageInfo); err != nil {
			t.Fatalf("invalid page-info JSON: %v", err)
		}
		if pageInfo.TotalPages < 1 || len(pageInfo.Pages) == 0 {
			t.Errorf("expected at least one page in page-info, got totalPages=%d pages=%d", pageInfo.TotalPages, len(pageInfo.Pages))
		}
	})

	t.Run("TC 10: Validación en /api/v1/redact/page-info con bundle corrupto (Fail)", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("pdf", "corrupted.pdf")
		part.Write([]byte("basura"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/v1/redact/page-info", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			t.Errorf("Expected failure status (400 or 500) for corrupt PDF, got %v", w.Code)
		}
	})
}
