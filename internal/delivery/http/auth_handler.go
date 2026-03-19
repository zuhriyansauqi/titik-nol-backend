package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

type AuthHandler struct {
	AuthUsecase domain.AuthUsecase
}

func NewAuthHandler(r *gin.Engine, authUsecase domain.AuthUsecase, authMiddleware gin.HandlerFunc) {
	handler := &AuthHandler{
		AuthUsecase: authUsecase,
	}

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/google", handler.LoginWithGoogle)
		authGroup.GET("/me", authMiddleware, handler.GetCurrentUser)
	}
}

func (h *AuthHandler) LoginWithGoogle(c *gin.Context) {
	var req domain.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	res, err := h.AuthUsecase.LoginWithGoogle(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Authentication failed", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", res)
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found in context", nil)
		return
	}

	user, err := h.AuthUsecase.GetCurrentUser(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		response.InternalServerError(c, "Failed to get user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User fetched successfully", user)
}

