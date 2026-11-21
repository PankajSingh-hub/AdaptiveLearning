package middleware

import (
	"net/http"
	"sync"
	"time"

	"adaptive-assessment/internal/utils"

	"github.com/gin-gonic/gin"
)

// IPLimiter allows at most `max` requests per IP in a sliding window. In-memory, per server instance.
type IPLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	max    int
	window time.Duration
}

func NewIPLimiter(max int, window time.Duration) *IPLimiter {
	return &IPLimiter{hits: map[string][]time.Time{}, max: max, window: window}
}

func (l *IPLimiter) Guard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, now := c.ClientIP(), time.Now()
		l.mu.Lock()
		if len(l.hits) > 5000 { // drop stale entries so the map cannot grow forever
			for k, v := range l.hits {
				if len(v) == 0 || now.Sub(v[len(v)-1]) > l.window {
					delete(l.hits, k)
				}
			}
		}
		recent := make([]time.Time, 0, len(l.hits[ip])+1)
		for _, t := range l.hits[ip] {
			if now.Sub(t) < l.window {
				recent = append(recent, t)
			}
		}
		if len(recent) >= l.max {
			l.hits[ip] = recent
			l.mu.Unlock()
			utils.Fail(c, http.StatusTooManyRequests, "too many requests, please try again later")
			c.Abort()
			return
		}
		l.hits[ip] = append(recent, now)
		l.mu.Unlock()
		c.Next()
	}
}