//nolint:revive // package comment
package main

import (
	"fmt"
	"log"
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
	if _, profiling := os.LookupEnv("ENABLE_PROFILING"); profiling {
		f, err := os.Create("/tmp/mem.prof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create memory profile: %v\n", err)
		} else {
			defer func() {
				if err := f.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "could not close memory profile: %v\n", err)
				}
			}()
			defer func() {
				fmt.Fprintln(os.Stderr, "Writing memory profile...")
				if err := pprof.WriteHeapProfile(f); err != nil {
					fmt.Fprintf(os.Stderr, "could not write memory profile: %v\n", err)
				}
				fmt.Fprintln(os.Stderr, "Memory profile written")
			}()
		}
	}

	// Ensure math fonts are available (downloads missing ones in background)
	go fontutils.EnsureMathFonts()

	// Use release mode to disable debug overhead
	gin.SetMode(gin.ReleaseMode)

	// gin.New() instead of gin.Default() — avoids the Logger middleware
	// which serializes stdout writes under a mutex on every request.
	router := gin.New()

	// Lightweight custom recovery: only captures stack on actual panic
	// (gin.Recovery() has per-request overhead from defer/stack-trace setup)
	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "[Recovery] panic recovered: %v\n", r)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	})

	// Only add request logger in debug mode (GIN_MODE=debug)
	if gin.Mode() == gin.DebugMode {
		router.Use(gin.Logger())
	}
	// Concurrency control: match GOMAXPROCS to minimize context switching
	// for CPU-bound PDF generation workloads.
	maxConcurrent := runtime.GOMAXPROCS(0)
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	semaphore := make(chan struct{}, maxConcurrent)
	fmt.Printf("Server starting with %d max concurrent workers\n", maxConcurrent)

	router.Use(func(c *gin.Context) {
		select {
		case semaphore <- struct{}{}:
		case <-c.Request.Context().Done():
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
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

	// PERF-43/151: never log.Fatal from a goroutine; surface listen errors on a channel
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		log.Printf("listen: %s\n", err)
		os.Exit(1)
	case <-quit:
		log.Println("Shutting down server...")
	}
}
