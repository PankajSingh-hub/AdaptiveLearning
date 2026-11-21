package middleware

import (
	"net/http"
	"strings"

	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

// BodyLimit caps request bodies before any JSON is parsed. Code submissions get a larger cap.
func BodyLimit(defaultBytes, codingBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := defaultBytes
		if strings.Contains(c.Request.URL.Path, "/coding/") {
			limit = codingBytes
		}
		if c.Request.ContentLength > limit {
			utils.Fail(c, http.StatusRequestEntityTooLarge, "request body is too large")
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}