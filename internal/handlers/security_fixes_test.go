package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/handlers/mocks"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

// stubServicePolicy gives the mock the real body-limit/error policy so tests
// exercise production 413/422 semantics without per-test expectations.
func stubServicePolicy(mockSvc *mocks.MockPDFService) {
	mockSvc.EXPECT().ReadUpload(gomock.Any(), gomock.Any()).DoAndReturn(
		func(r io.Reader, kind string) ([]byte, bool, error) {
			return readBounded(r, uploadLimitFor(kind))
		}).AnyTimes()
	mockSvc.EXPECT().UploadLimit(gomock.Any()).DoAndReturn(uploadLimitFor).AnyTimes()
	mockSvc.EXPECT().ClassifyError(gomock.Any()).DoAndReturn(pdfErrorStatus).AnyTimes()
}

func setupSecurityRouter(t *testing.T) (*gin.Engine, *mocks.MockPDFService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockPDFService(ctrl)
	stubServicePolicy(mockSvc)
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
	r.POST("/api/v1/compress", handleCompressPDF)
	r.GET("/api/v1/fonts", handleGetFonts)
	r.POST("/api/v1/fonts", handleUploadFont)
	r.POST("/api/v1/htmltopdf", handleHTMLToPDF)
	r.POST("/api/v1/htmltoimage", handleHTMLToImage)
	r.POST("/api/v1/redact/page-info", handleRedactPageInfo)
	r.POST("/api/v1/redact/apply", handleRedactApply)
	return r, mockSvc
}

func multipartFileBody(t *testing.T, field, name string, content []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(field, name)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part failed: %v", err)
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("WriteField failed: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("multipart close failed: %v", err)
	}
	return body, w.FormDataContentType()
}

func TestReadBounded(t *testing.T) {
	if _, ok, err := readBounded(bytes.NewReader([]byte("abc")), 3); err != nil || !ok {
		t.Fatalf("exact-limit read should pass: ok=%v err=%v", ok, err)
	}
	if _, ok, err := readBounded(bytes.NewReader([]byte("abcd")), 3); err != nil || ok {
		t.Fatalf("over-limit read should fail closed: ok=%v err=%v", ok, err)
	}
}

func TestHandleMergePDFs_OversizedUpload413(t *testing.T) {
	r, _ := setupSecurityRouter(t)
	big := make([]byte, maxPDFBytes+1)
	body, ctype := multipartFileBody(t, "pdf", "a.pdf", big, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/merge", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want 413 body=%s", w.Code, w.Body.String())
	}
}

func TestHandleFillPDF_OversizedUpload413(t *testing.T) {
	r, _ := setupSecurityRouter(t)
	big := make([]byte, maxPDFBytes+1)
	body, ctype := multipartFileBody(t, "pdf", "form.pdf", big, nil)

	// Oversized pdf alone must 413 before xfdf presence is even considered.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fill", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want 413", w.Code)
	}
}

func TestHandlerSplitPDF_OversizedUpload413(t *testing.T) {
	r, _ := setupSecurityRouter(t)
	big := make([]byte, maxPDFBytes+1)
	body, ctype := multipartFileBody(t, "pdf", "doc.pdf", big, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/split", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want 413", w.Code)
	}
}

func TestHandleRedactPageInfo_OversizedUpload413(t *testing.T) {
	r, _ := setupSecurityRouter(t)
	big := make([]byte, maxPDFBytes+1)
	body, ctype := multipartFileBody(t, "pdf", "doc.pdf", big, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/redact/page-info", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want 413", w.Code)
	}
}

func TestHandleUploadFont_Oversized413(t *testing.T) {
	r, _ := setupSecurityRouter(t)
	big := make([]byte, maxFontBytes+1)
	body, ctype := multipartFileBody(t, "font", "big.ttf", big, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fonts", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want 413 body=%s", w.Code, w.Body.String())
	}
}

func TestHandleGenerateTemplatePDF_OversizedJSON413(t *testing.T) {
	r, _ := setupSecurityRouter(t)
	big := bytes.Repeat([]byte("a"), maxTemplateJSONBody+1)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewReader(big))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want 413", w.Code)
	}
}

func TestHandleGenerateTemplatePDF_InvalidJSONGeneric(t *testing.T) {
	r, _ := setupSecurityRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want 400", w.Code)
	}
	if got := w.Body.String(); got != `{"code":"invalid_input","error":"invalid template data","message":"invalid template data"}` {
		t.Fatalf("leaky body: %q", got)
	}
}

func TestHandleMergePDFs_MalformedMapsTo422(t *testing.T) {
	r, mockSvc := setupSecurityRouter(t)
	mockSvc.EXPECT().MergePDFs(gomock.Any()).Return(nil, errors.New("invalid xref table: corrupt file"))

	body, ctype := multipartFileBody(t, "pdf", "a.pdf", []byte("%PDF-broken"), nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/merge", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d want 422 body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid PDF input") {
		t.Fatalf("expected generic message, got %s", w.Body.String())
	}
}

func TestHandleMergePDFs_EngineFailureMapsTo500(t *testing.T) {
	r, mockSvc := setupSecurityRouter(t)
	mockSvc.EXPECT().MergePDFs(gomock.Any()).Return(nil, errors.New("worker pool crashed"))

	body, ctype := multipartFileBody(t, "pdf", "a.pdf", []byte("%PDF-a"), nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/merge", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d want 500 body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "worker pool crashed") {
		t.Fatalf("backend detail leaked: %s", w.Body.String())
	}
}

func TestValidateFetchURL(t *testing.T) {
	blocked := []string{
		"http://127.0.0.1/x",
		"http://10.0.0.1/",
		"http://172.16.0.5/",
		"http://192.168.1.1/",
		"http://169.254.169.254/latest/meta-data/",
		"http://[::1]/",
		"http://localhost/",
		"http://example.localhost/",
	}
	for _, raw := range blocked {
		if err := validateFetchURL(context.Background(), raw); !errors.Is(err, errFetchURLBlocked) {
			t.Fatalf("url %q should be blocked, got %v", raw, err)
		}
	}
	invalid := []string{"ftp://example.com/x", "notaurl", "http://"}
	for _, raw := range invalid {
		if err := validateFetchURL(context.Background(), raw); err == nil || errors.Is(err, errFetchURLBlocked) {
			t.Fatalf("url %q should be invalid, got %v", raw, err)
		}
	}
	if err := validateFetchURL(context.Background(), ""); err != nil {
		t.Fatalf("empty url (html path) should pass, got %v", err)
	}
	if err := validateFetchURL(context.Background(), "http://93.184.216.1/"); err != nil {
		t.Fatalf("public IP literal should pass, got %v", err)
	}
}

type recordingFetchURLResolver struct {
	ctx context.Context
}

func (r *recordingFetchURLResolver) LookupIP(ctx context.Context, _, _ string) ([]net.IP, error) {
	r.ctx = ctx
	return []net.IP{net.ParseIP("93.184.216.1")}, nil
}

func TestValidateFetchURLUsesRequestContext(t *testing.T) {
	resolver := &recordingFetchURLResolver{}
	previous := defaultFetchURLResolver
	defaultFetchURLResolver = resolver
	t.Cleanup(func() { defaultFetchURLResolver = previous })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := validateFetchURL(ctx, "http://example.test/"); err != nil {
		t.Fatalf("validateFetchURL: %v", err)
	}
	if resolver.ctx != ctx {
		t.Fatal("DNS lookup did not receive the request context")
	}
}

func TestHandleHTMLToImageCanonicalizesFormatForMIME(t *testing.T) {
	r, mockSvc := setupSecurityRouter(t)
	mockSvc.EXPECT().HTMLToImage(gomock.Any()).DoAndReturn(func(req models.HTMLToImageRequest) ([]byte, error) {
		if req.Format != "jpg" {
			t.Fatalf("format passed to service = %q, want jpg", req.Format)
		}
		return []byte("jpeg bytes"), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/htmltoimage", bytes.NewBufferString(`{"html":"<h1>ok</h1>","format":" JPEG "}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("content type = %q, want image/jpeg", got)
	}
	if got := w.Header().Get("Content-Disposition"); got != "attachment; filename=converted.jpg" {
		t.Fatalf("content disposition = %q, want canonical jpg filename", got)
	}
}

func TestHandleHTMLToPDF_Validation(t *testing.T) {
	r, mockSvc := setupSecurityRouter(t)
	mockSvc.EXPECT().HTMLToPDF(gomock.Any()).Return([]byte("%PDF-html"), nil).Times(1)

	cases := []struct {
		name string
		body string
		code int
	}{
		{"neither html nor url", `{}`, http.StatusBadRequest},
		{"bad scheme", `{"url":"ftp://example.com/x"}`, http.StatusBadRequest},
		{"loopback blocked", `{"url":"http://127.0.0.1/x"}`, http.StatusForbidden},
		{"metadata blocked", `{"url":"http://169.254.169.254/"}`, http.StatusForbidden},
		{"valid html", `{"html":"<h1>hi</h1>"}`, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/htmltopdf", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("status=%d want %d body=%s", w.Code, tc.code, w.Body.String())
			}
		})
	}
}

func TestHandleHTMLToPDF_OversizedBody413(t *testing.T) {
	r, _ := setupSecurityRouter(t)
	big := append([]byte(`{"html":"`), bytes.Repeat([]byte("a"), maxHTMLBodyBytes)...)
	big = append(big, []byte(`"}`)...)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/htmltopdf", bytes.NewReader(big))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want 413", w.Code)
	}
}

func TestHandleHTMLToImage_MissingBoth400(t *testing.T) {
	r, _ := setupSecurityRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/htmltoimage", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want 400 body=%s", w.Code, w.Body.String())
	}
}

func TestHandleGetFonts_MockSuccess(t *testing.T) {
	r, mockSvc := setupSecurityRouter(t)
	mockSvc.EXPECT().GetFonts().Return([]models.FontInfo{{ID: "roboto", Name: "Roboto"}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fonts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "roboto") {
		t.Fatalf("missing font in body: %s", w.Body.String())
	}
}

func TestHandleUploadFont_MockSuccess(t *testing.T) {
	r, mockSvc := setupSecurityRouter(t)
	mockSvc.EXPECT().RegisterFont("Test", gomock.Any()).Return(nil)

	body, ctype := multipartFileBody(t, "font", "Test.ttf", []byte("fakefontdata"), nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fonts", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestFastAPIKeepsAuthMiddleware(t *testing.T) {
	t.Setenv("GIN_FAST_API", "1")
	t.Setenv("REQUIRE_AUTH", "1")
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	RegisterRoutes(engine)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want 401: fast path must keep auth enforcement", w.Code)
	}
}

func TestPprofRejectsSpoofedForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router)

	// Real loopback peer passes.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	router.ServeHTTP(w, req)
	if w.Code == http.StatusForbidden {
		t.Fatal("loopback peer unexpectedly forbidden")
	}

	// Spoofed X-Forwarded-For with a public peer must be forbidden.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	req.RemoteAddr = "203.0.113.7:1234"
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("spoofed forwarded-for not forbidden: got %d", w.Code)
	}
}
