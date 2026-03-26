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

// SetupAccounts godoc
// @Summary      Setup initial accounts
// @Description  Create initial bank accounts during user onboarding
// @Tags         onboarding
// @Accept       json
// @Produce      json
// @Param        request body domain.SetupAccountsRequest true "Setup Accounts Data"
// @Success      201  {object}  response.Response{data=[]domain.Account}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/onboarding/accounts [post]
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
