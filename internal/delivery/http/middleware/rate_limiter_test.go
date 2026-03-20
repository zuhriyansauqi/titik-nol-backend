package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/delivery/http/middleware"
	"github.com/stretchr/testify/assert"
)

func setupRateLimiterRouter(rps float64, burst int) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.RateLimiter(rps, burst))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRateLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	r := setupRateLimiterRouter(10, 5)

	// 5 requests should pass (the burst limit)
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimiter_BlocksRequestsOverLimit(t *testing.T) {
	r := setupRateLimiterRouter(10, 5)

	// First 5 requests pass (burst)
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:1234" // different IP
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 6th request should be blocked immediately
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Rate limit exceeded")
}

func TestRateLimiter_BypassesIfRateLimitingIsZero(t *testing.T) {
	r := setupRateLimiterRouter(0, 0)

	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.3:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimiter_DifferentIPsHaveDifferentLimits(t *testing.T) {
	r := setupRateLimiterRouter(1, 1)

	// First IP
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.4:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	
	req1Blocked, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req1Blocked.RemoteAddr = "192.168.1.4:1234"
	w1Blocked := httptest.NewRecorder()
	r.ServeHTTP(w1Blocked, req1Blocked)
	assert.Equal(t, http.StatusTooManyRequests, w1Blocked.Code)

	// Second IP should still pass
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.5:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
