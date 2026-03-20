package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

type DashboardHandler struct {
	dashboardUsecase domain.DashboardUsecase
}

func NewDashboardHandler(rg *gin.RouterGroup, uc domain.DashboardUsecase) {
	handler := &DashboardHandler{dashboardUsecase: uc}
	rg.GET("/dashboard", handler.GetSummary)
}

func (h *DashboardHandler) GetSummary(c *gin.Context) {
	userID, _ := c.Get("user_id")

	summary, err := h.dashboardUsecase.GetSummary(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Dashboard summary fetched successfully", summary)
}
