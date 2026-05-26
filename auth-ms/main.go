// Command auth-ms is a small, self-contained authentication microservice:
// email/password accounts in SQLite, bcrypt hashing, HS256 JWT issuance.
package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	dbPath := env("AUTH_DB_PATH", "auth.db")
	secret := env("AUTH_JWT_SECRET", "dev-insecure-secret-change-me")
	corsOrigin := env("AUTH_CORS_ORIGIN", "*")
	addr := ":" + env("AUTH_PORT", "9090")

	store, err := OpenStore(dbPath)
	if err != nil {
		log.Fatalf("auth-ms: open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	tm := NewTokenManager(secret, 24*time.Hour)
	srv := newServer(store, tm, corsOrigin)

	log.Printf("auth-ms listening on %s (db=%s)", addr, dbPath)
	if err := srv.Run(addr); err != nil {
		log.Fatalf("auth-ms: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
