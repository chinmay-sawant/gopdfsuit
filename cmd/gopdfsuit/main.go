package main

import (
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
	if os.Getenv("DISABLE_PROFILING") == "" {
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
	// Only Recovery() is kept for crash safety.
	router := gin.New()
	router.Use(gin.Recovery())

	// Only add request logger in debug mode (GIN_MODE=debug)
	if gin.Mode() == gin.DebugMode {
		router.Use(gin.Logger())
	}
	// Concurrency control: match to CPU count to minimize context switching
	// for CPU-bound PDF generation workloads
	// maxConcurrent := runtime.NumCPU()
	// Concurrency control: 100 allows high utilization without overwhelming the system
	// for the 100 VU benchmark targets.
	maxConcurrent := 100
	semaphore := make(chan struct{}, maxConcurrent)
	log.Printf("Server starting with %d max concurrent workers (CPUs: %d)", maxConcurrent, runtime.NumCPU())

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
