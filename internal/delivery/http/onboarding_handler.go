package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

type OnboardingHandler struct {
	onboardingUsecase domain.OnboardingUsecase
}

func NewOnboardingHandler(rg *gin.RouterGroup, uc domain.OnboardingUsecase) {
	handler := &OnboardingHandler{onboardingUsecase: uc}
	rg.POST("/onboarding/accounts", handler.SetupAccounts)
}

func (h *OnboardingHandler) SetupAccounts(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.SetupAccountsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	result, err := h.onboardingUsecase.SetupAccounts(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Accounts setup successfully", result)
}
