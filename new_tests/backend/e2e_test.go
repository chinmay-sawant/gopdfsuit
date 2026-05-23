package backend_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func findRepoRoot(t *testing.T) string {
	t.Helper()

	startDir, err := os.Getwd()
	require.NoError(t, err)

	currentDir := startDir
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			t.Fatalf("could not find repository root from %s", startDir)
		}
		currentDir = parentDir
	}
}

func startE2EServer(t *testing.T) (baseURL string, shutdown func()) {
	t.Helper()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	})

	semaphore := make(chan struct{}, 48)
	router.Use(func(c *gin.Context) {
		semaphore <- struct{}{}
		defer func() {
			<-semaphore
		}()
		c.Next()
	})

	handlers.RegisterRoutes(router)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &http.Server{
		Handler:     router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Serve(listener)
	}()

	baseURL = "http://" + listener.Addr().String()

	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(10 * time.Second)
	for {
		resp, err := client.Get(baseURL + "/")
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not become ready at %s: %v", baseURL, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)

		select {
		case err := <-serverErr:
			if err != nil && err != http.ErrServerClosed {
				t.Fatalf("server stopped with error: %v", err)
			}
		default:
		}
	}

	return baseURL, shutdown
}

func TestE2ETemplatePDFFlow(t *testing.T) {
	repoRoot := findRepoRoot(t)
	baseURL, shutdown := startE2EServer(t)
	defer shutdown()

	redirectClient := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := redirectClient.Get(baseURL + "/")
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
	require.Equal(t, "/gopdfsuit", resp.Header.Get("Location"))

	spaResp, err := http.Get(baseURL + "/gopdfsuit")
	require.NoError(t, err)
	defer func() {
		_ = spaResp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, spaResp.StatusCode)
	spaBody, err := io.ReadAll(spaResp.Body)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(string(spaBody)), "<!doctype html")

	payloadPath := filepath.Join(repoRoot, "sampledata", "editor", "financial_digitalsignature.json")
	payload, err := os.ReadFile(payloadPath)
	require.NoError(t, err)

	pdfResp, err := http.Post(baseURL+"/api/v1/generate/template-pdf", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer func() {
		_ = pdfResp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, pdfResp.StatusCode)
	require.Equal(t, "application/pdf", pdfResp.Header.Get("Content-Type"))
	require.Contains(t, pdfResp.Header.Get("Content-Disposition"), "generated.pdf")

	pdfBody, err := io.ReadAll(pdfResp.Body)
	require.NoError(t, err)
	require.Greater(t, len(pdfBody), 1000)
	require.True(t, bytes.HasPrefix(pdfBody, []byte("%PDF-")), "response must be a PDF document")
}