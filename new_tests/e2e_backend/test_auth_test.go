package e2e_backend_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
		t.Skip("Skipping: K_SERVICE not set (not in Cloud Run mode). Run with: K_SERVICE=e2e-test go test ./new_tests/e2e_backend/")
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
		t.Skip("Skipping: K_SERVICE not set (not in Cloud Run mode). Run with: K_SERVICE=e2e-test go test ./new_tests/e2e_backend/")
	}

	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/test/auth", nil)
	req.Header.Set("Authorization", "Bearer invalid-test-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid")
}

func TestTestAuthEndpoint_CloudRunSubprocess(t *testing.T) {
	if os.Getenv("K_SERVICE") != "" {
		t.Skip("Running inside K_SERVICE subprocess; skip to avoid recursion")
	}

	cmd := exec.Command("go", "test", "-v", "-run", "TestTestAuthEndpoint_CloudRun", "github.com/chinmay-sawant/gopdfsuit/v5/new_tests/e2e_backend")
	cmd.Env = append(os.Environ(),
		"K_SERVICE=e2e-test",
		"GOOGLE_CLIENT_ID=46981518442-ap2ga76ao0sj82t47mimcv8eot0l0pbt.apps.googleusercontent.com",
	)
	output, err := cmd.CombinedOutput()
	t.Logf("Subprocess output:\n%s", string(output))
	assert.NoError(t, err, "Cloud Run auth tests failed in subprocess")
}
