package handlers

import (
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

// decodeJSONBody caps the request body at limit bytes and unmarshals it into
// T. It is the single JSON-decode policy for the HTML conversion handlers
// (the pooled generate path in decode.go stays as its override). tooLarge
// reports http.MaxBytesReader rejections so callers emit 413; any other
// failure is a 400 with the response already unwritten (caller aborts).
func decodeJSONBody[T any](c *gin.Context, limit int64) (req T, tooLarge bool, err error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)

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
