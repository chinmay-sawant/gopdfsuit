package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// panicRecoveryHandler is a named handler so gin registers a stable function
// pointer instead of allocating a new closure per router.Use call.
func panicRecoveryHandler(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Recovery] panic recovered: %v", r)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}()
	c.Next()
}

// PanicRecovery returns lightweight panic recovery middleware.
// Unlike gin.Recovery(), it does not capture stack traces on panic.
func PanicRecovery() gin.HandlerFunc {
	return panicRecoveryHandler
}