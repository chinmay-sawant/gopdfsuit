package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleRedactApply_TextSearchWorksViaMultipart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/redact/apply", HandleRedactApply)

	pdfPath := filepath.Join("..", "..", "sampledata", "financialreport", "template.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("failed to read sample PDF %s: %v", pdfPath, err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	fw, err := mw.CreateFormFile("pdf", "template.pdf")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := fw.Write(pdfBytes); err != nil {
		t.Fatalf("failed writing PDF payload: %v", err)
	}

	if err := mw.WriteField("blocks", "[]"); err != nil {
		t.Fatalf("WriteField blocks failed: %v", err)
	}
	if err := mw.WriteField("mode", "secure_required"); err != nil {
		t.Fatalf("WriteField mode failed: %v", err)
	}
	if err := mw.WriteField("textSearch", `[{"text":"SECTION"},{"text":"COM"}]`); err != nil {
		t.Fatalf("WriteField textSearch failed: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("multipart close failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/redact/apply", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Fatalf("expected application/pdf response, got %q", ct)
	}

	if bytes.Equal(w.Body.Bytes(), pdfBytes) {
		t.Fatal("expected redacted output to differ from original PDF")
	}

	reportHeader := w.Header().Get("X-Redaction-Report")
	if reportHeader == "" {
		t.Fatal("expected X-Redaction-Report header to be present")
	}
	var report map[string]any
	if err := json.Unmarshal([]byte(reportHeader), &report); err != nil {
		t.Fatalf("failed to parse report header: %v", err)
	}
	generated, ok := report["generatedRects"].(float64)
	if !ok || generated <= 0 {
		t.Fatalf("expected generatedRects > 0 in report, got: %v", report["generatedRects"])
	}

	// Store output at repository root for easier inspection by developer
	outputPath := filepath.Join("..", "..", "template_redacted_web.pdf")
	if err := os.WriteFile(outputPath, w.Body.Bytes(), 0o600); err != nil {
		t.Fatalf("failed to write redacted output PDF: %v", err)
	}
	t.Logf("redacted test output written to: %s", outputPath)
}

func TestHandleRedactApply_RejectsEmptyPDF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/redact/apply", HandleRedactApply)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("pdf", "empty.pdf")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := fw.Write(nil); err != nil {
		t.Fatalf("failed writing empty payload: %v", err)
	}
	_ = mw.WriteField("mode", "secure_required")
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/redact/apply", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty pdf, got %d", w.Code)
	}
}
