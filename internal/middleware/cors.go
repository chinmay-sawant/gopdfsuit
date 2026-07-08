package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Predeclared CORS header values (PERF-201: avoid repeated map growth / string churn).
const (
	corsAllowOrigin  = "https://chinmay-sawant.github.io"
	corsAllowHeaders = "*"
	corsAllowMethods = "*"
	corsExpose       = "X-Redaction-Report"
)

// CORSMiddleware handles CORS headers and preflight requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", corsAllowOrigin)
		h.Set("Access-Control-Allow-Headers", corsAllowHeaders)
		h.Set("Access-Control-Allow-Methods", corsAllowMethods)
		h.Set("Access-Control-Expose-Headers", corsExpose)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
