package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
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

	var mu sync.Mutex
	limiters := make(map[string]*struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	})

	go func() {
		for {
			time.Sleep(3 * time.Minute)
			mu.Lock()
			for ip, cl := range limiters {
				if time.Since(cl.lastSeen) > 3*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		cl, exists := limiters[ip]
		if !exists {
			cl = &struct {
				limiter  *rate.Limiter
				lastSeen time.Time
			}{
				limiter: rate.NewLimiter(rate.Limit(rps), burst),
			}
			limiters[ip] = cl
		}
		cl.lastSeen = time.Now()

		return cl.limiter
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			slog.WarnContext(c.Request.Context(), "Rate limit exceeded", "network.client.ip", ip)
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "Please try again later.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
