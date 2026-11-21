package middleware

import (
	"errors"
	"log"
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
		claims, err := auth.Authenticate(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			var ae *service.AppError
			if errors.As(err, &ae) {
				utils.Fail(c, ae.Status, ae.Message)
			} else {
				log.Printf("auth lookup failed: %v", err)
				utils.Fail(c, http.StatusInternalServerError, "something went wrong, please try again")
			}
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
