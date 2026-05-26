package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func handleRegister(store *Store, tm *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username (min 3 chars) and password (min 8 chars) required"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}

		user, err := store.CreateUser(strings.TrimSpace(req.Email), string(hash))
		if errors.Is(err, ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}

		token, err := tm.Issue(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}
		c.JSON(http.StatusCreated, tokenResponse{Token: token, User: user})
	}
}

func handleLogin(store *Store, tm *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
			return
		}

		user, hash, err := store.Credentials(strings.TrimSpace(req.Email))
		// Same response for unknown user and wrong password (no account enumeration).
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := tm.Issue(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}
		c.JSON(http.StatusOK, tokenResponse{Token: token, User: user})
	}
}

func handleVerify(tm *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": "Authorization: Bearer <token> required"})
			return
		}
		claims, err := tm.Verify(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": "invalid token"})
			return
		}
		c.JSON(http.StatusOK, verifyResponse{Valid: true, Sub: claims.Subject, Email: claims.Email})
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
