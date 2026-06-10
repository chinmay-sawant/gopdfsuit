//nolint:revive // package comment
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers"
	"github.com/chinmay-sawant/gopdfsuit/v5/pkg/fontutils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Profiling is opt-in to avoid heap instrumentation overhead in production/benchmarks
	if os.Getenv("ENABLE_PROFILING") == "1" {
		f, err := os.Create("/tmp/mem.prof")
		if err == nil {
			defer f.Close()
			defer func() { _ = pprof.WriteHeapProfile(f) }()
		}
	}

	// Ensure math fonts are available (downloads missing ones in background)
	go fontutils.EnsureMathFonts()

	// Use release mode to disable debug overhead
	gin.SetMode(gin.ReleaseMode)

	// gin.New() instead of gin.Default() — avoids the Logger middleware
	// which serializes stdout writes under a mutex on every request.
	router := gin.New()

	router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	// Concurrency control: match to CPU count to minimize context switching
	// for CPU-bound PDF generation workloads.
	// Using NumCPU() prevents goroutine thrashing — 100 goroutines on 24 cores
	// caused massive context-switch overhead and was the primary bottleneck.
	maxConcurrent := runtime.NumCPU()
	semaphore := make(chan struct{}, maxConcurrent)
	router.Use(func(c *gin.Context) {
		semaphore <- struct{}{}
		defer func() { <-semaphore }()
		c.Next()
	})

	handlers.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "listen: %s\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	os.Stderr.WriteString("Shutting down server...\n")
}
