package middleware

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter creates a middleware that limits requests by client IP
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	// If rps is 0, disable rate limiting entirely
	if rps <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	var mu sync.RWMutex
	limiters := make(map[string]*rate.Limiter)

	getLimiter := func(ip string) *rate.Limiter {
		mu.RLock()
		limiter, exists := limiters[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			// Double check after acquiring write lock
			limiter, exists = limiters[ip]
			if !exists {
				limiter = rate.NewLimiter(rate.Limit(rps), burst)
				limiters[ip] = limiter
			}
			mu.Unlock()
		}

		return limiter
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			slog.Warn("Rate limit exceeded", "ip", ip)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
