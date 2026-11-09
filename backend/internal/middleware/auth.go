package middleware

import (
	"net/http"
	"strings"

	"adaptive-assessment/internal/service"
	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			utils.Fail(c, http.StatusUnauthorized, "missing authorization token")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			utils.Fail(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole is what stops a candidate from calling admin APIs (403).
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			utils.Fail(c, http.StatusForbidden, "you do not have permission to access this resource")
			c.Abort()
			return
		}
		c.Next()
	}
}