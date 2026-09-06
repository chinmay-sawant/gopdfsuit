//nolint:revive // package comment
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/handlers"
	"github.com/chinmay-sawant/gopdfsuit/v7/pkg/fontutils"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := handlers.ResolveServerConfig()
	defer handlers.MaybeEnableProfiling()()
	// Ensure math fonts are available (downloads missing ones in background)
	go fontutils.EnsureMathFonts()
	handlers.WarmRuntime()

	// Use release mode to disable debug overhead
	gin.SetMode(gin.ReleaseMode)

	// gin.New() instead of gin.Default() - avoids the Logger middleware
	// which serializes stdout writes under a mutex on every request.
	router := handlers.NewRouter(cfg)
	srv := handlers.NewServer(cfg, router)

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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "forced shutdown: %s\n", err)
	}
}
