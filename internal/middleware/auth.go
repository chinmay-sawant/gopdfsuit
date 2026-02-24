// Package middleware provides HTTP middlewares for the application.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

// isCloudRunCached is evaluated once at package init to avoid per-request os.Getenv overhead.
var isCloudRunCached = os.Getenv("K_SERVICE") != "" || os.Getenv("K_REVISION") != ""

// IsCloudRun checks if the application is running on Google Cloud Run
func IsCloudRun() bool {
	return isCloudRunCached
}

// GoogleAuthMiddleware validates Google OAuth ID tokens
// Only enforces authentication when running on Cloud Run
func GoogleAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if not running on Cloud Run
		if !IsCloudRun() {
			c.Next()
			return
		}

		// Skip authentication for OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Get the expected audience (your Cloud Run service URL)
		// This should be set as an environment variable
		audience := os.Getenv("GOOGLE_OAUTH_AUDIENCE")
		if audience == "" {
			// Try Client ID as audience (common for Google Sign-In)
			audience = os.Getenv("GOOGLE_CLIENT_ID")
		}
		if audience == "" {
			// If not set, try to get from Cloud Run metadata
			audience = os.Getenv("CLOUD_RUN_SERVICE_URL")
		}

		// Validate the ID token
		ctx := context.Background()
		payload, err := idtoken.Validate(ctx, token, audience)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid ID token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Token is valid, store user info in context
		c.Set("user_email", payload.Claims["email"])
		c.Set("user_name", payload.Claims["name"])
		c.Set("user_picture", payload.Claims["picture"])
		c.Set("user_sub", payload.Subject)

		c.Next()
	}
}

// OptionalAuthMiddleware checks for authentication but doesn't enforce it
// Useful for endpoints that can work with or without auth
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if not on Cloud Run
		if !IsCloudRun() {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth provided, continue without user info
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		audience := os.Getenv("GOOGLE_OAUTH_AUDIENCE")
		if audience == "" {
			audience = os.Getenv("GOOGLE_CLIENT_ID")
		}
		if audience == "" {
			audience = os.Getenv("CLOUD_RUN_SERVICE_URL")
		}

		ctx := context.Background()
		payload, err := idtoken.Validate(ctx, token, audience)
		if err == nil {
			// Token is valid, store user info
			c.Set("user_email", payload.Claims["email"])
			c.Set("user_name", payload.Claims["name"])
			c.Set("user_picture", payload.Claims["picture"])
			c.Set("user_sub", payload.Subject)
		}

		c.Next()
	}
}

// GetUserEmail retrieves the authenticated user's email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}

// GetUserInfo retrieves all user info from context
func GetUserInfo(c *gin.Context) map[string]interface{} {
	userInfo := make(map[string]interface{})

	if email, exists := c.Get("user_email"); exists {
		userInfo["email"] = email
	}
	if name, exists := c.Get("user_name"); exists {
		userInfo["name"] = name
	}
	if picture, exists := c.Get("user_picture"); exists {
		userInfo["picture"] = picture
	}
	if sub, exists := c.Get("user_sub"); exists {
		userInfo["sub"] = sub
	}

	return userInfo
}

// LogAuthInfo logs authentication information (useful for debugging)
func LogAuthInfo(c *gin.Context) {
	if IsCloudRun() {
		userInfo := GetUserInfo(c)
		if len(userInfo) > 0 {
			fmt.Printf("Authenticated user: %+v\n", userInfo)
		} else {
			fmt.Println("No authenticated user")
		}
	}
}
