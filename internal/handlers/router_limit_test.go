package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

type countingReader struct {
	reader io.Reader
	read   int
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.read += n
	return n, err
}

// TestConcurrencyLimiterShedsLoad asserts a full limiter returns 429 with
// Retry-After: 1 instead of queueing, and the depth gauge drains (C3).
func TestConcurrencyLimiterShedsLoad(t *testing.T) {
	gin.SetMode(gin.TestMode)
	blocker := make(chan struct{})
	unblock := make(chan struct{})
	r := gin.New()
	r.Use(concurrencyLimiter(1))
	r.GET("/x", func(c *gin.Context) {
		close(blocker)
		<-unblock
		c.Status(http.StatusOK)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		r.ServeHTTP(w, req)
	}()
	<-blocker
	if d := ConcurrencyDepth(); d != 1 {
		close(unblock)
		wg.Wait()
		t.Fatalf("depth while blocked = %d, want 1", d)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		close(unblock)
		wg.Wait()
		t.Fatalf("overflow code = %d, want 429", w2.Code)
	}
	if w2.Header().Get("Retry-After") != "1" {
		close(unblock)
		wg.Wait()
		t.Fatalf("Retry-After = %q, want 1", w2.Header().Get("Retry-After"))
	}
	close(unblock)
	wg.Wait()
	if d := ConcurrencyDepth(); d != 0 {
		t.Fatalf("depth after drain = %d, want 0", d)
	}
}

func TestRoutePolicyLimitsMultipartBodyBeforeParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	payload := bytes.Repeat([]byte("x"), 4096)
	counted := &countingReader{reader: bytes.NewReader(payload)}
	router := NewRouter(ServerConfig{
		MaxConcurrent: 1,
		Policy: RoutePolicy{
			MaxMergeBodyBytes: 128,
			MaxMergeFiles:     2,
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/merge", io.NopCloser(counted))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=not-used")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", w.Code)
	}
	if counted.read >= len(payload) {
		t.Fatalf("body reader consumed all %d bytes before rejection", counted.read)
	}
}
