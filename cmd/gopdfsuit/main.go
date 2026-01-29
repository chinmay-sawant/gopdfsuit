package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/chinmay-sawant/gopdfsuit/internal/handlers"
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

	router := gin.Default()
	handlers.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		// service connections
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
