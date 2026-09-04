package handlers

import (
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/middleware"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
	"github.com/gin-gonic/gin"
)

// ServerConfig is the composition-root input for the HTTP server. It is
// resolved once at startup from flags/env (ResolveServerConfig) and then
// threaded through NewRouter / NewServer so main only declares the wiring.
type ServerConfig struct {
	Addr            string
	MaxConcurrent   int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	EnableProfiling bool
}

// ResolveServerConfig reads MAX_CONCURRENT / BENCH_MODE / ENABLE_PROFILING
// from the environment. Behavior matches the previous inline main logic:
// explicit MAX_CONCURRENT wins, BENCH_MODE scales with CPU capped at 48,
// otherwise NumCPU.
func ResolveServerConfig() ServerConfig {
	return ServerConfig{
		Addr:            ":8080",
		MaxConcurrent:   resolveMaxConcurrent(),
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    60 * time.Second,
		ShutdownTimeout: 15 * time.Second,
		EnableProfiling: os.Getenv("ENABLE_PROFILING") == "1",
	}
}

func resolveMaxConcurrent() int {
	if v := os.Getenv("MAX_CONCURRENT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	if os.Getenv("BENCH_MODE") == "1" {
		n := runtime.NumCPU() * 2
		if n > 48 {
			return 48
		}
		if n < 1 {
			return 1
		}
		return n
	}
	return runtime.NumCPU()
}

// WarmRuntime pre-warms engine pools and the JSON decode schema so the first
// request pays no cold-start cost. Font provisioning (fontutils) stays with
// the main composition root, which owns the background download goroutine.
func WarmRuntime() {
	pdf.WarmRuntimePools()
	WarmJSONDecode()
}

// MaybeEnableProfiling starts the opt-in heap-profile dump (ENABLE_PROFILING=1)
// and returns a cleanup func. It is opt-in to avoid heap instrumentation
// overhead in production/benchmarks.
func MaybeEnableProfiling() func() {
	noop := func() {}
	if os.Getenv("ENABLE_PROFILING") != "1" {
		return noop
	}
	f, err := os.Create("/tmp/mem.prof")
	if err != nil {
		return noop
	}
	return func() {
		defer f.Close()
		_ = pprof.WriteHeapProfile(f)
	}
}

// concurrencyDepth tracks in-flight requests admitted by concurrencyLimiter.
// Exported via ConcurrencyDepth for load-tuning observability (C3/B4).
var concurrencyDepth atomic.Int64

// ConcurrencyDepth reports current in-flight admitted requests.
func ConcurrencyDepth() int64 {
	return concurrencyDepth.Load()
}

// concurrencyLimiter caps in-flight requests with a semaphore channel.
// Try-acquire is non-blocking: when full it aborts 429 with Retry-After: 1
// instead of queueing, so overload sheds fast and stays observable.
// SetLimit (page-compress errgroup) and signWorkerSlots are separate
// inner budgets and intentionally untouched (see plans/adrs ADR note).
func concurrencyLimiter(n int) gin.HandlerFunc {
	semaphore := make(chan struct{}, n)
	return func(c *gin.Context) {
		select {
		case semaphore <- struct{}{}:
		default:
			c.Header("Retry-After", "1")
			abortErrorAndStop(c, http.StatusTooManyRequests, "server busy")
			return
		}
		concurrencyDepth.Add(1)
		defer func() {
			<-semaphore
			concurrencyDepth.Add(-1)
		}()
		c.Next()
	}
}

// NewRouter composes the middleware chain (recovery, concurrency limit) and
// the API route table. Auth/CORS behavior lives in RegisterRoutes and the
// middleware package and is unchanged.
func NewRouter(cfg ServerConfig) *gin.Engine {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(gin.CustomRecovery(func(c *gin.Context, _ any) {
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	router.Use(concurrencyLimiter(cfg.MaxConcurrent))
	RegisterRoutes(router)
	return router
}

// NewServer builds the HTTP server with fixed timeouts from cfg.
func NewServer(cfg ServerConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}
