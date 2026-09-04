package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareAssignsAndEchoes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		if GetRequestID(c) == "" {
			t.Error("expected request id in context")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
}

func TestRequestIDMiddlewareHonorsIncoming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		if got := GetRequestID(c); got != "caller-123" {
			t.Errorf("request id = %q, want caller-123", got)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", "caller-123")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("X-Request-ID"); got != "caller-123" {
		t.Errorf("echoed header = %q, want caller-123", got)
	}
}
