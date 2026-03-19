package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attributes := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user-agent", c.Request.UserAgent()),
			// Note: request_id is automatically injected by ContextHandler via context
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				slog.LogAttrs(c.Request.Context(), slog.LevelError, e, attributes...)
			}
		} else {
			slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "request", attributes...)
		}
	}
}
