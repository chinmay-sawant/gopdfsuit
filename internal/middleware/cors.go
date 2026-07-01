package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS headers and preflight requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "https://chinmay-sawant.github.io")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Expose-Headers", "X-Redaction-Report")

		c.Next()
	}
}
