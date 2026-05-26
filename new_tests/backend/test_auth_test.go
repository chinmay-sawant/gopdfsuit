package backend_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// mintToken signs an auth-ms-style HS256 JWT with the given shared secret.
func mintToken(secret, email string) string {
	claims := jwt.MapClaims{
		"email": email,
		"sub":   "1",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	s, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return s
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	handlers.RegisterRoutes(r)
	return r
}

func TestTestAuthEndpoint_LocalMode(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/test/auth", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "authenticated")
}

func TestTestAuthEndpoint_CloudRunRequiresAuth(t *testing.T) {
	if os.Getenv("K_SERVICE") == "" {
		t.Skip("Skipping: K_SERVICE not set (not in Cloud Run mode). Run with: K_SERVICE=e2e-test go test ./new_tests/backend/")
	}

	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/test/auth", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization")
}

func TestTestAuthEndpoint_CloudRunInvalidToken(t *testing.T) {
	if os.Getenv("K_SERVICE") == "" {
		t.Skip("Skipping: K_SERVICE not set (not in Cloud Run mode). Run with: K_SERVICE=e2e-test go test ./new_tests/backend/")
	}

	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/test/auth", nil)
	req.Header.Set("Authorization", "Bearer invalid-test-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid")
}

func TestTestAuthEndpoint_CloudRunValidToken(t *testing.T) {
	if os.Getenv("K_SERVICE") == "" {
		t.Skip("Skipping: K_SERVICE not set (not in Cloud Run mode). Run with: K_SERVICE=e2e-test go test ./new_tests/backend/")
	}

	secret := os.Getenv("AUTH_JWT_SECRET")
	token := mintToken(secret, "tester@example.com")

	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/test/auth", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "authenticated")
	assert.Contains(t, w.Body.String(), "tester@example.com")
}

func TestTestAuthEndpoint_CloudRunSubprocess(t *testing.T) {
	if os.Getenv("K_SERVICE") != "" {
		t.Skip("Running inside K_SERVICE subprocess; skip to avoid recursion")
	}

	cmd := exec.Command("go", "test", "-v", "-run", "TestTestAuthEndpoint_CloudRun", "github.com/chinmay-sawant/gopdfsuit/v5/new_tests/backend")
	cmd.Env = append(os.Environ(),
		"K_SERVICE=e2e-test",
		"AUTH_JWT_SECRET=e2e-test-secret",
	)
	output, err := cmd.CombinedOutput()
	t.Logf("Subprocess output:\n%s", string(output))
	assert.NoError(t, err, "Cloud Run auth tests failed in subprocess")
}
