package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	handler "github.com/mzhryns/titik-nol-backend/internal/delivery/http"
)

func setupAuthRouter(mockUC *mocks.MockAuthUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No real auth middleware for handler tests — we test handlers in isolation
	handler.NewAuthHandler(r, mockUC, func(c *gin.Context) { c.Next() })
	return r
}

func TestLoginWithGoogle_Success(t *testing.T) {
	mockUC := new(mocks.MockAuthUsecase)
	r := setupAuthRouter(mockUC)

	reqBody := domain.GoogleLoginRequest{IDToken: "valid-token"}
	body, _ := json.Marshal(reqBody)

	authResp := &domain.AuthResponse{AccessToken: "jwt-token", IsNewUser: true}
	mockUC.On("LoginWithGoogle", mock.Anything, mock.AnythingOfType("*domain.GoogleLoginRequest")).Return(authResp, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/google", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Login successful", res.Message)
}

func TestLoginWithGoogle_BadBody(t *testing.T) {
	mockUC := new(mocks.MockAuthUsecase)
	r := setupAuthRouter(mockUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/google", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginWithGoogle_AuthError(t *testing.T) {
	mockUC := new(mocks.MockAuthUsecase)
	r := setupAuthRouter(mockUC)

	reqBody := domain.GoogleLoginRequest{IDToken: "bad-token"}
	body, _ := json.Marshal(reqBody)

	mockUC.On("LoginWithGoogle", mock.Anything, mock.AnythingOfType("*domain.GoogleLoginRequest")).Return(nil, domain.ErrInvalidCredentials)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/google", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUser_Success(t *testing.T) {
	mockUC := new(mocks.MockAuthUsecase)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	userID := uuid.New()
	expectedUser := &domain.User{ID: userID, Email: "test@example.com", Name: "Test"}

	// Middleware that sets user_id (simulating auth middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	handler.NewAuthHandler(r, mockUC, func(c *gin.Context) { c.Next() })

	mockUC.On("GetCurrentUser", mock.Anything, userID).Return(expectedUser, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
}

func TestGetCurrentUser_Unauthorized(t *testing.T) {
	mockUC := new(mocks.MockAuthUsecase)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No user_id middleware — simulates missing auth
	handler.NewAuthHandler(r, mockUC, func(c *gin.Context) { c.Next() })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
