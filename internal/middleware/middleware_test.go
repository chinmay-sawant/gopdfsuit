package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGoogleAuthMiddleware_RequireAuthEnforces(t *testing.T) {
	t.Setenv("REQUIRE_AUTH", "1")
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(GoogleAuthMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want 401", w.Code)
	}
	if strings.Contains(w.Body.String(), "details") {
		t.Fatalf("error detail leaked: %s", w.Body.String())
	}
}

func TestGoogleAuthMiddleware_OpenByDefault(t *testing.T) {
	if IsCloudRun() {
		t.Skip("running on Cloud Run, open-by-default does not apply")
	}
	t.Setenv("REQUIRE_AUTH", "")
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(GoogleAuthMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want 200 (open local default)", w.Code)
	}
}

func TestCORSMiddleware_Enumerated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, PATCH, DELETE, OPTIONS" {
		t.Fatalf("methods=%q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); strings.Contains(got, "*") {
		t.Fatalf("headers must be enumerated, got %q", got)
	}
	if !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Fatalf("headers=%q", w.Header().Get("Access-Control-Allow-Headers"))
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("vary=%q", got)
	}
}
