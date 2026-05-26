// Package middleware provides HTTP middlewares for the application.
package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// isCloudRunCached is evaluated once at package init to avoid per-request os.Getenv overhead.
var isCloudRunCached = os.Getenv("K_SERVICE") != "" || os.Getenv("K_REVISION") != ""

// authEnabledCached lets us enforce auth anywhere (e.g. local dev / make run),
// independent of the Cloud Run deploy platform.
var authEnabledCached = os.Getenv("AUTH_ENABLED") == "true"

// IsCloudRun checks if the application is running on Cloud Run
func IsCloudRun() bool {
	return isCloudRunCached
}

// authRequired reports whether requests must carry a valid auth-ms JWT.
func authRequired() bool {
	return authEnabledCached || isCloudRunCached
}

// authSecret is the shared HS256 secret used to verify JWTs minted by auth-ms.
// It must match AUTH_JWT_SECRET configured on the auth-ms service.
func authSecret() []byte {
	s := os.Getenv("AUTH_JWT_SECRET")
	if s == "" {
		s = "dev-insecure-secret-change-me"
	}
	return []byte(s)
}

type authClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func verifyToken(tokenString string) (*authClaims, error) {
	claims := &authClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return authSecret(), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// AuthMiddleware validates auth-ms JWTs and only enforces authentication when
// running on Cloud Run (mirrors the frontend, which gates the UI there).
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authRequired() {
			c.Next()
			return
		}
		// Skip authentication for OPTIONS requests (CORS preflight)
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		token, ok := bearerToken(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		claims, err := verifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		setUserContext(c, claims)
		c.Next()
	}
}

// OptionalAuthMiddleware attaches user info when a valid token is present but
// never rejects the request. Useful for endpoints that work with or without auth.
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authRequired() {
			c.Next()
			return
		}
		token, ok := bearerToken(c)
		if !ok {
			c.Next()
			return
		}
		if claims, err := verifyToken(token); err == nil {
			setUserContext(c, claims)
		}
		c.Next()
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func setUserContext(c *gin.Context, claims *authClaims) {
	c.Set("user_email", claims.Email)
	c.Set("user_sub", claims.Subject)
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
	if sub, exists := c.Get("user_sub"); exists {
		userInfo["sub"] = sub
	}

	return userInfo
}
