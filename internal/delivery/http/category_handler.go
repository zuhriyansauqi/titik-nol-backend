package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

type CategoryHandler struct {
	categoryUsecase domain.CategoryUsecase
}

func NewCategoryHandler(rg *gin.RouterGroup, uc domain.CategoryUsecase) {
	handler := &CategoryHandler{categoryUsecase: uc}
	rg.POST("/categories", handler.BulkCreate)
	rg.GET("/categories", handler.Fetch)
}

// BulkCreate godoc
// @Summary      Create multiple categories
// @Description  Create multiple expense or income categories at once
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        request body domain.BulkCreateCategoryRequest true "Categories Data"
// @Success      201  {object}  response.Response{data=[]domain.Category}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/categories [post]
func (h *CategoryHandler) BulkCreate(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.BulkCreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	categories, err := h.categoryUsecase.BulkCreate(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Categories created successfully", categories)
}

// Fetch godoc
// @Summary      Get all categories
// @Description  Fetch all categories belonging to the authenticated user
// @Tags         categories
// @Produce      json
// @Param        type   query     string  false  "Filter by category type (expense/income)"
// @Success      200    {object}  response.Response{data=[]domain.Category}
// @Failure      500    {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/categories [get]
func (h *CategoryHandler) Fetch(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filterType *domain.CategoryType
	if t := c.Query("type"); t != "" {
		ct := domain.CategoryType(t)
		filterType = &ct
	}

	categories, err := h.categoryUsecase.FetchByUserID(c.Request.Context(), userID.(uuid.UUID), filterType)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Categories fetched successfully", categories)
}
