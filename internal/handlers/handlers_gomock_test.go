package handlers

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers/mocks"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf/merge"
	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func setupMockRouter(t *testing.T) (*gin.Engine, *mocks.MockPDFService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockPDFService(ctrl)
	SetPDFService(mockSvc)
	t.Cleanup(func() {
		SetPDFService(nil)
		ctrl.Finish()
	})

	r := gin.New()
	r.POST("/api/v1/generate/template-pdf", handleGenerateTemplatePDF)
	r.POST("/api/v1/fill", handleFillPDF)
	r.POST("/api/v1/merge", handleMergePDFs)
	r.POST("/api/v1/split", handlerSplitPDF)
	return r, mockSvc
}

func TestHandleGenerateTemplatePDF_MockSuccess(t *testing.T) {
	r, mockSvc := setupMockRouter(t)
	mockSvc.EXPECT().
		GenerateTemplatePDF(gomock.Any()).
		Return([]byte("%PDF-mock"), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewBufferString(`{"config":{"page":"A4","pageAlignment":1}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != mimeTypePDF {
		t.Fatalf("content-type=%q", ct)
	}
	if !bytes.Equal(w.Body.Bytes(), []byte("%PDF-mock")) {
		t.Fatalf("unexpected body: %q", w.Body.Bytes())
	}
}

func TestHandleGenerateTemplatePDF_MockServiceError(t *testing.T) {
	r, mockSvc := setupMockRouter(t)
	mockSvc.EXPECT().
		GenerateTemplatePDF(gomock.Any()).
		Return(nil, errors.New("generation failed"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewBufferString(`{"config":{"page":"A4","pageAlignment":1}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandleGenerateTemplatePDF_InvalidJSON(t *testing.T) {
	r, _ := setupMockRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandleFillPDF_MockSuccess(t *testing.T) {
	r, mockSvc := setupMockRouter(t)
	mockSvc.EXPECT().
		FillPDFWithXFDF(gomock.Any(), gomock.Any()).
		DoAndReturn(func(pdfBytes, xfdfBytes []byte) ([]byte, error) {
			if len(pdfBytes) == 0 || len(xfdfBytes) == 0 {
				t.Fatal("expected non-empty pdf and xfdf")
			}
			return []byte("%PDF-filled"), nil
		})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	pdfPart, _ := writer.CreateFormFile("pdf", "form.pdf")
	_, _ = pdfPart.Write([]byte("%PDF-1.4"))
	xfdfPart, _ := writer.CreateFormFile("xfdf", "data.xfdf")
	_, _ = xfdfPart.Write([]byte(`<?xml version="1.0"?><xfdf/>`))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fill", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandleFillPDF_MissingFields(t *testing.T) {
	r, _ := setupMockRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fill", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandleMergePDFs_MockSuccess(t *testing.T) {
	r, mockSvc := setupMockRouter(t)
	mockSvc.EXPECT().
		MergePDFs(gomock.Any()).
		Return([]byte("%PDF-merged"), nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("pdf", "a.pdf")
	_, _ = part.Write([]byte("%PDF-a"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/merge", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandleSplitPDF_MockSingleOutput(t *testing.T) {
	r, mockSvc := setupMockRouter(t)
	mockSvc.EXPECT().
		SplitPDF(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ []byte, spec merge.SplitSpec) ([][]byte, error) {
			_ = spec
			return [][]byte{[]byte("%PDF-split")}, nil
		})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("pdf", "doc.pdf")
	_, _ = part.Write([]byte("%PDF-src"))
	_ = writer.WriteField("pages", "1")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/split", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != mimeTypePDF {
		t.Fatalf("content-type=%q", ct)
	}
}
