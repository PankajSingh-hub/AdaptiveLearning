package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{"success": true, "message": message, "data": data})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success":   false,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
		"path":      c.Request.URL.Path,
	})
}