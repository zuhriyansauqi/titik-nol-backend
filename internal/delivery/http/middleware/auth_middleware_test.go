package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/delivery/http/middleware"
	jwtmocks "github.com/mzhryns/titik-nol-backend/internal/pkg/jwt/mocks"
	"github.com/stretchr/testify/assert"
)

func setupMiddlewareRouter(mockJWT *jwtmocks.MockJWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AuthMiddleware(mockJWT))
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
	return r
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	mockJWT := new(jwtmocks.MockJWTService)
	r := setupMiddlewareRouter(mockJWT)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_BadFormat(t *testing.T) {
	mockJWT := new(jwtmocks.MockJWTService)
	r := setupMiddlewareRouter(mockJWT)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic some-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockJWT := new(jwtmocks.MockJWTService)
	r := setupMiddlewareRouter(mockJWT)

	mockJWT.On("ValidateToken", "bad-token").Return(uuid.Nil, errors.New("invalid"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockJWT := new(jwtmocks.MockJWTService)
	r := setupMiddlewareRouter(mockJWT)

	userID := uuid.New()
	mockJWT.On("ValidateToken", "valid-token").Return(userID, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
