package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestV1GroupAlwaysCarriesAuthMiddleware pins the route-policy invariant:
// the v1 group runs GoogleAuthMiddleware whether or not the GIN_FAST_API=1
// benchmark path sheds CORS. With REQUIRE_AUTH=1 an unauthenticated POST
// must get 401 in both modes; a 4xx/5xx from deeper in the stack would mean
// auth was skipped.
func TestV1GroupAlwaysCarriesAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, fastAPI := range []string{"", "1"} {
		t.Run("GIN_FAST_API="+fastAPI, func(t *testing.T) {
			t.Setenv("GIN_FAST_API", fastAPI)
			t.Setenv("REQUIRE_AUTH", "1")
			t.Setenv("K_SERVICE", "")
			t.Setenv("K_REVISION", "")
			t.Setenv("GOOGLE_OAUTH_AUDIENCE", "")
			t.Setenv("GOOGLE_CLIENT_ID", "")
			t.Setenv("CLOUD_RUN_SERVICE_URL", "")

			cfg := ResolveServerConfig()
			if want := fastAPI != "1"; cfg.Policy.EnableCORS != want {
				t.Fatalf("EnableCORS = %v, want %v", cfg.Policy.EnableCORS, want)
			}
			if !cfg.Policy.RequireAuth {
				t.Fatal("RequireAuth = false with REQUIRE_AUTH=1")
			}

			router := NewRouter(cfg)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/compress", nil)
			router.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("GIN_FAST_API=%q: status = %d, want 401 (auth middleware skipped)", fastAPI, w.Code)
			}
		})
	}
}
