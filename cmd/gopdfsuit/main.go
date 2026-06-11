//nolint:revive // package comment
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"syscall"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/handlers"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf"
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
	pdf.WarmRuntimePools()
	handlers.WarmJSONDecode()

	// Use release mode to disable debug overhead
	gin.SetMode(gin.ReleaseMode)

	// gin.New() instead of gin.Default() — avoids the Logger middleware
	// which serializes stdout writes under a mutex on every request.
	router := gin.New()

	router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	maxConcurrent := resolveMaxConcurrent()
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
