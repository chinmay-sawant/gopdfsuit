package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// newServer wires the auth routes onto a fresh gin engine. Tests build it with
// an in-memory store; main builds it with a file-backed store.
func newServer(store *Store, tm *TokenManager, corsOrigin string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(corsOrigin))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	auth.POST("/register", handleRegister(store, tm))
	auth.POST("/login", handleLogin(store, tm))
	auth.GET("/verify", handleVerify(tm))

	return r
}

// corsMiddleware allows the frontend origin to call the API. Tokens travel in
// the Authorization header (not cookies), so wildcard origin is acceptable here.
func corsMiddleware(origin string) gin.HandlerFunc {
	if origin == "" {
		origin = "*"
	}
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
