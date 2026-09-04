package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestIDKey is the gin context key carrying the request ID.
const RequestIDKey = "request_id"

// RequestIDMiddleware assigns every request an ID (honoring an incoming
// X-Request-ID), echoes it back on the response, and emits exactly one
// structured line per request at the handler boundary. Per-handler error
// logs keep backend detail; this line carries only the observable fields.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = newRequestID()
		}
		c.Set(RequestIDKey, rid)
		c.Header("X-Request-ID", rid)
		start := time.Now()
		c.Next()
		log.Printf("request request_id=%s method=%s path=%s status=%d latency=%s",
			rid, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start).Round(time.Microsecond))
	}
}

// GetRequestID returns the ID assigned by RequestIDMiddleware, or "" when
// the middleware did not run (e.g. unit tests wiring handlers directly).
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b[:])
}
