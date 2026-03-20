package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/logger"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/jwt"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

func AuthMiddleware(jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "Unauthorized", "Authorization header is required", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, 401, "Unauthorized", "Authorization header must be in the format 'Bearer <token>'", nil)
			c.Abort()
			return
		}

		userID, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			response.Error(c, 401, "Unauthorized", "Invalid or expired token", nil)
			c.Abort()
			return
		}

		c.Set("user_id", userID)

		// Propagate user_id into request context so all downstream slog calls
		// automatically include it via ContextHandler.
		ctx := context.WithValue(c.Request.Context(), logger.UserIDContextKey, userID.String())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
