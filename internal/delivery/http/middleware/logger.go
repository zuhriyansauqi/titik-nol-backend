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

		// Use Datadog-friendly attribute names for automatic facet mapping.
		// See: https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/
		attributes := []slog.Attr{
			slog.Int("http.status_code", status),
			slog.String("http.method", c.Request.Method),
			slog.String("http.url_details.path", path),
			slog.String("http.url_details.queryString", query),
			slog.String("network.client.ip", c.ClientIP()),
			slog.Duration("duration", latency),
			slog.Int("http.response_content_length", c.Writer.Size()),
			slog.String("http.useragent", c.Request.UserAgent()),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				attributes = append(attributes,
					slog.String("error.message", e),
					slog.String("error.kind", "gin_error"),
				)
				slog.LogAttrs(c.Request.Context(), slog.LevelError, "request_error", attributes...)
			}
		} else if status >= 500 {
			slog.LogAttrs(c.Request.Context(), slog.LevelError, "request", attributes...)
		} else if status >= 400 {
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "request", attributes...)
		} else {
			slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "request", attributes...)
		}
	}
}
