package handlers

import (
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

const (
	bodyLimitContextKey      = "gopdfsuit.body-limit"
	mergeFileLimitContextKey = "gopdfsuit.merge-file-limit"
)

func applyBodyLimit(c *gin.Context, limit int64) bool {
	if limit <= 0 {
		return true
	}
	if c.Request.ContentLength > limit {
		abortError(c, http.StatusRequestEntityTooLarge, "request body too large")
		return false
	}
	if _, exists := c.Get(bodyLimitContextKey); exists {
		return true
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	c.Set(bodyLimitContextKey, limit)
	return true
}

func requestBodyLimit(c *gin.Context, fallback int64) int64 {
	if value, exists := c.Get(bodyLimitContextKey); exists {
		if limit, ok := value.(int64); ok {
			return limit
		}
	}
	return fallback
}

func ensureMultipartBodyLimit(c *gin.Context) bool {
	return applyBodyLimit(c, requestBodyLimit(c, maxMultipartBodyBytes))
}

func mergeFileCountLimit(c *gin.Context) int {
	if value, exists := c.Get(mergeFileLimitContextKey); exists {
		if limit, ok := value.(int); ok {
			return limit
		}
	}
	return maxMergeFiles
}

// decodeJSONBody caps the request body at limit bytes and unmarshals it into
// T. It is the single JSON-decode policy for the HTML conversion handlers
// (the pooled generate path in decode.go stays as its override). tooLarge
// reports http.MaxBytesReader rejections so callers emit 413; any other
// failure is a 400 with the response already unwritten (caller aborts).
func decodeJSONBody[T any](c *gin.Context, limit int64) (req T, tooLarge bool, err error) {
	limit = requestBodyLimit(c, limit)
	if !applyBodyLimit(c, limit) {
		return req, true, http.ErrBodyReadAfterClose
	}

	data, rerr := c.GetRawData()
	if rerr != nil {
		if isBodyTooLargeErr(rerr) {
			return req, true, rerr
		}
		log.Printf("decodeJSONBody: read body failed: %v", rerr)
		return req, false, rerr
	}

	if uerr := sonic.Unmarshal(data, &req); uerr != nil {
		log.Printf("decodeJSONBody: invalid JSON: %v", uerr)
		return req, false, uerr
	}
	return req, false, nil
}
