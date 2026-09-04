package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestRouteTable pins the composition-root route table: the generate,
// compress, and one redact route must exist on the router NewRouter builds,
// so renames/deletions fail loudly here instead of as 404s in production.
func TestRouteTable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(ServerConfig{MaxConcurrent: 4})

	have := map[string]bool{}
	for _, r := range router.Routes() {
		have[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{
		"POST /api/v1/generate/template-pdf",
		"POST /api/v1/compress",
		"POST /api/v1/redact/apply",
	} {
		if !have[want] {
			t.Fatalf("route %q missing from router table", want)
		}
	}
}

// TestConcurrencyLimiterShedsAndReadmits pins limiter behavior through the
// shed path: with one permit held, overflow gets 429 + Retry-After: 1, and
// after the holder drains a new request is admitted (no stuck semaphore).
func TestConcurrencyLimiterShedsAndReadmits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entered := make(chan struct{}, 10)
	release := make(chan struct{})
	r := gin.New()
	r.Use(concurrencyLimiter(1))
	r.GET("/slow", func(c *gin.Context) {
		entered <- struct{}{}
		<-release
		c.Status(http.StatusOK)
	})

	type result struct {
		code int
		hdr  http.Header
	}
	serve := func() result {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/slow", nil))
		return result{w.Code, w.Header()}
	}
	doneFirst := make(chan result, 1)
	go func() { doneFirst <- serve() }()

	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("first request never entered the handler")
	}

	doneSecond := make(chan result, 1)
	go func() { doneSecond <- serve() }()

	select {
	case r2 := <-doneSecond:
		if r2.code != http.StatusTooManyRequests {
			close(release)
			t.Fatalf("overflow status = %d, want 429", r2.code)
		}
		if r2.hdr.Get("Retry-After") != "1" {
			close(release)
			t.Fatalf("Retry-After = %q, want 1", r2.hdr.Get("Retry-After"))
		}
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("overflow request never resolved")
	}

	// Drain the holder (release is closed, so the handler proceeds): the
	// permit is freed and a new request is admitted and served.
	close(release)
	select {
	case r1 := <-doneFirst:
		if r1.code != http.StatusOK {
			t.Fatalf("first status = %d, want 200", r1.code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("first request never completed after release")
	}

	doneThird := make(chan result, 1)
	go func() { doneThird <- serve() }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("third request never entered after drain")
	}
	select {
	case r3 := <-doneThird:
		if r3.code != http.StatusOK {
			t.Fatalf("third status = %d, want 200", r3.code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("third request never completed after drain")
	}
}
