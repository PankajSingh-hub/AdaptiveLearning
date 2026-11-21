package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimit(10, 100))
	read := func(c *gin.Context) { _, _ = io.ReadAll(c.Request.Body); c.Status(http.StatusOK) }
	r.POST("/api/x", read)
	r.POST("/api/coding/y", read)

	do := func(path string, n int) int {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(strings.Repeat("a", n))))
		return w.Code
	}
	if do("/api/x", 5) != http.StatusOK {
		t.Fatal("a small body must pass")
	}
	if do("/api/x", 50) != http.StatusRequestEntityTooLarge {
		t.Fatal("a body over the default limit must be rejected")
	}
	if do("/api/coding/y", 50) != http.StatusOK {
		t.Fatal("coding routes allow a larger body")
	}
	if do("/api/coding/y", 500) != http.StatusRequestEntityTooLarge {
		t.Fatal("a body over the coding limit must be rejected")
	}
}