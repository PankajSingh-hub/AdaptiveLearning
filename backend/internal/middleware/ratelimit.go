package middleware

import (
	"net/http"
	"sync"
	"time"

	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

type attempt struct {
	fails int
	reset time.Time
}

// LoginLimiter blocks an IP after too many failed logins. It is in-memory, so it is
// per server instance; behind several instances you would use Redis instead.
type LoginLimiter struct {
	mu       sync.Mutex
	byIP     map[string]*attempt
	maxFails int
	window   time.Duration
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{byIP: map[string]*attempt{}, maxFails: 5, window: 15 * time.Minute}
}

func (l *LoginLimiter) Guard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l.mu.Lock()
		a := l.byIP[ip]
		if a != nil && time.Now().After(a.reset) {
			delete(l.byIP, ip)
			a = nil
		}
		blocked := a != nil && a.fails >= l.maxFails
		l.mu.Unlock()
		if blocked {
			utils.Fail(c, http.StatusTooManyRequests, "too many failed attempts, try again in a few minutes")
			c.Abort()
			return
		}

		c.Next()

		l.mu.Lock()
		defer l.mu.Unlock()
		switch c.Writer.Status() {
		case http.StatusUnauthorized:
			if l.byIP[ip] == nil {
				l.byIP[ip] = &attempt{reset: time.Now().Add(l.window)}
			}
			l.byIP[ip].fails++
		case http.StatusOK:
			delete(l.byIP, ip)
		}
	}
}