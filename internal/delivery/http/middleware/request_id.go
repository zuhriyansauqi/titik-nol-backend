package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/logger"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set(string(logger.RequestContextKey), requestID)

		// Create a new context with the request ID and set it to the request
		ctx := context.WithValue(c.Request.Context(), logger.RequestContextKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
