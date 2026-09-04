package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
	"github.com/gin-gonic/gin"
)

// TestClassifyErrorSentinels pins the errors.Is/As primary path: taxonomy
// errors map to their codes without relying on message substrings.
func TestClassifyErrorSentinels(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{fmt.Errorf("wrap: %w", gopdflib.ErrInvalidInput), http.StatusUnprocessableEntity},
		{fmt.Errorf("wrap: %w", gopdflib.ErrLimitExceeded), http.StatusRequestEntityTooLarge},
		{fmt.Errorf("wrap: %w", gopdflib.ErrUpstream), http.StatusBadGateway},
		{fmt.Errorf("wrap: %w", font.ErrUpstream), http.StatusBadGateway},
		{&font.HTTPStatusError{URL: "https://example.com/f.ttf", Status: 503}, http.StatusBadGateway},
		{fmt.Errorf("wrap: %w", gopdflib.ErrInternal), http.StatusInternalServerError},
		// Pinned legacy fallback for foreign errors with only message text.
		{errors.New("invalid xref table: corrupt file"), http.StatusUnprocessableEntity},
		{errors.New("worker pool crashed"), http.StatusInternalServerError},
		{errors.New("PDF exceeds maximum size (33554432 bytes)"), http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		if got := pdfErrorStatus(tc.err); got != tc.want {
			t.Errorf("pdfErrorStatus(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}

// TestLimitSubstringConsistent pins the "limit" drift fix: an over-cap
// foreign error with only message text maps to 413 through the shared
// gopdflib.ClassifyMessage source, the same list wrapEngineError uses.
func TestLimitSubstringConsistent(t *testing.T) {
	for _, msg := range []string{
		"output hit the page limit",
		"PDF exceeds maximum size (33554432 bytes)",
	} {
		if got := pdfErrorStatus(errors.New(msg)); got != http.StatusRequestEntityTooLarge {
			t.Errorf("pdfErrorStatus(%q) = %d, want 413", msg, got)
		}
		if code := gopdflib.ClassifyMessage(errors.New(msg)); code != gopdflib.CodeLimitExceeded {
			t.Errorf("ClassifyMessage(%q) = %q, want limit_exceeded", msg, code)
		}
	}
}

// TestAbortPDFErrorFallbackMessage pins the fallbackMsg parameter: an
// unclassified backend failure replies 500 with the caller's message, so
// redaction keeps its "redaction failed" text without a parallel helper.
func TestAbortPDFErrorFallbackMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/redact/apply", nil)

	abortPDFError(c, errors.New("worker pool crashed"), "redaction failed")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	want := `{"code":"internal","error":"redaction failed","message":"redaction failed"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

// TestPprofForbiddenEnvelope pins the shared errorBody shape on pprof denials.
func TestPprofForbiddenEnvelope(t *testing.T) {
	if pprofForbiddenResp["error"] != "Forbidden: Pprof is only accessible from localhost" {
		t.Fatalf("error = %q", pprofForbiddenResp["error"])
	}
	if pprofForbiddenResp["message"] != pprofForbiddenResp["error"] {
		t.Fatalf("message %q != error alias %q", pprofForbiddenResp["message"], pprofForbiddenResp["error"])
	}
	if _, ok := pprofForbiddenResp["code"]; !ok {
		t.Fatal("pprof response missing envelope code")
	}
}

// TestAbortPDFErrorEnvelope pins the {code,message} body shape plus the
// legacy error alias on the central abort path.
func TestAbortPDFErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/merge", nil)

	abortPDFError(c, fmt.Errorf("merge: %w: corrupt", gopdflib.ErrInvalidInput), "PDF processing failed")
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", w.Code)
	}
	want := `{"code":"invalid_input","error":"invalid PDF input","message":"invalid PDF input"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

// TestAbortPDFErrorUpstream pins the 502 path for upstream failures.
func TestAbortPDFErrorUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/merge", nil)

	abortPDFError(c, &font.HTTPStatusError{URL: "https://example.com/f.ttf", Status: 503}, "PDF processing failed")
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", w.Code)
	}
	want := `{"code":"upstream","error":"upstream dependency failed","message":"upstream dependency failed"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}
