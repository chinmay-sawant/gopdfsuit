package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/handlers/mocks"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf"
	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

// TestGenerateBorrowedSeam proves the FastGenerateService branch serves the
// pooled path: the mock implements the seam, so the handler must call it
// instead of the copying GenerateTemplatePDF fallback.
func TestGenerateBorrowedSeam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockPDFService(ctrl)
	stubServicePolicy(mockSvc)
	fast := &mocks.FastMockPDFService{MockPDFService: mockSvc}
	called := false
	fast.BorrowedFunc = func(template models.PDFTemplate) (*pdf.BorrowedPDF, error) {
		called = true
		if template.Config.Page != "A4" {
			t.Fatalf("template page = %q, want A4", template.Config.Page)
		}
		return &pdf.BorrowedPDF{}, nil
	}
	SetPDFService(fast)
	t.Cleanup(func() {
		SetPDFService(nil)
		ctrl.Finish()
	})

	r := gin.New()
	r.POST("/api/v1/generate/template-pdf", handleGenerateTemplatePDF)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf",
		bytes.NewBufferString(`{"config":{"page":"A4","pageAlignment":1},"title":{"props":"a","text":"b"},"footer":{"font":"a","text":"b"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !called {
		t.Fatal("borrowed seam was not called")
	}
}
