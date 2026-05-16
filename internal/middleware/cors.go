package middleware

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// resolveAllowOrigin sets Access-Control-Allow-Origin for browser cross-origin calls.
// - GOPDFSUIT_CORS_ALLOW_ORIGIN: if set, always used (single origin string).
// - Cloud Run (default): GitHub Pages demo origin.
// - Elsewhere (local Docker, go run): reflect http(s)://localhost[:port] and 127.0.0.1, else "*".
func resolveAllowOrigin(c *gin.Context) string {
	if custom := strings.TrimSpace(os.Getenv("GOPDFSUIT_CORS_ALLOW_ORIGIN")); custom != "" {
		return custom
	}
	if IsCloudRun() {
		return "https://chinmay-sawant.github.io"
	}
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		return "*"
	}
	u, err := url.Parse(origin)
	if err != nil {
		return "*"
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return origin
	}
	return "*"
}

// CORSMiddleware handles CORS headers and preflight requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", resolveAllowOrigin(c))
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Expose-Headers", "X-Redaction-Report")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}