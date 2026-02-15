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

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Profiling is opt-in to avoid heap instrumentation overhead in production/benchmarks
	if os.Getenv("ENABLE_PROFILING") == "1" {
		f, err := os.Create("/tmp/mem.prof")
		if err != nil {
			log.Printf("could not create memory profile: %v", err)
		} else {
			defer func() {
				if err := f.Close(); err != nil {
					log.Printf("could not close memory profile: %v", err)
				}
			}()
			defer func() {
				log.Println("Writing memory profile...")
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Printf("could not write memory profile: %v", err)
				}
				log.Println("Memory profile written")
			}()
		}
	}

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
				log.Printf("[Recovery] panic recovered: %v", r)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	})

	// Only add request logger in debug mode (GIN_MODE=debug)
	if gin.Mode() == gin.DebugMode {
		router.Use(gin.Logger())
	}
	// Concurrency control: match to CPU count to minimize context switching
	// for CPU-bound PDF generation workloads.
	// Using NumCPU() prevents goroutine thrashing — 100 goroutines on 24 cores
	// caused massive context-switch overhead and was the primary bottleneck.
	// maxConcurrent := runtime.NumCPU()
	maxConcurrent := 48
	semaphore := make(chan struct{}, maxConcurrent)
	fmt.Printf("Server starting with %d max concurrent workers (CPUs: %d)\n", maxConcurrent, runtime.NumCPU())

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
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}
