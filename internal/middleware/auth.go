// Package middleware provides HTTP middlewares for the application.
package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

// IsCloudRun checks if the application is running on Google Cloud Run.
// It reads the environment on every call: Cloud Run sets K_SERVICE or
// K_REVISION for the life of the instance, so there is nothing to cache,
// and a package-init snapshot would freeze test overrides (t.Setenv) and
// any late-arriving environment.
func IsCloudRun() bool {
	return os.Getenv("K_SERVICE") != "" || os.Getenv("K_REVISION") != ""
}

// authEnforced reports whether GoogleAuthMiddleware must validate tokens.
//
// Open-by-default local behavior: authentication is enforced ONLY when this
// returns true, i.e. when running on Cloud Run (K_SERVICE or K_REVISION is
// set) or when REQUIRE_AUTH=1 is set in the environment. Everywhere else
// (local dev, plain Docker, unit tests) requests pass through untouched.
// Set REQUIRE_AUTH=1 to force enforcement outside Cloud Run (CI, staging,
// on-prem).
func authEnforced() bool {
	if os.Getenv("REQUIRE_AUTH") == "1" {
		return true
	}
	return IsCloudRun()
}

// resolveAudience returns the expected Google OAuth audience:
// GOOGLE_OAUTH_AUDIENCE, else GOOGLE_CLIENT_ID, else CLOUD_RUN_SERVICE_URL.
// Shared by the enforcing and optional middlewares so the precedence lives
// in one place.
func resolveAudience() string {
	if audience := os.Getenv("GOOGLE_OAUTH_AUDIENCE"); audience != "" {
		return audience
	}
	if audience := os.Getenv("GOOGLE_CLIENT_ID"); audience != "" {
		return audience
	}
	return os.Getenv("CLOUD_RUN_SERVICE_URL")
}

// validateToken validates a Google OAuth ID token against audience. It is
// the single seam over idtoken.Validate for both middlewares (and tests).
func validateToken(ctx context.Context, token, audience string) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, token, audience)
}

// GoogleAuthMiddleware validates Google OAuth ID tokens.
//
// Open-by-default local behavior: when authEnforced() is false (not on Cloud
// Run and REQUIRE_AUTH != "1") the middleware is a no-op and every request
// passes through. Set REQUIRE_AUTH=1 to default-deny outside Cloud Run.
func GoogleAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if not running on Cloud Run (unless REQUIRE_AUTH=1)
		if !authEnforced() {
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

		// Validate the ID token against the shared audience precedence.
		ctx := c.Request.Context()
		payload, err := validateToken(ctx, token, resolveAudience())
		if err != nil {
			log.Printf("GoogleAuthMiddleware: token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication failed",
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
		ctx := c.Request.Context()
		payload, err := validateToken(ctx, token, resolveAudience())
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
func GetUserInfo(c *gin.Context) map[string]any {
	userInfo := make(map[string]any)

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
