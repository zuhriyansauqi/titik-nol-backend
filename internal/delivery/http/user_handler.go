package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
	"gorm.io/gorm"
)

type UserHandler struct {
	UserUsecase domain.UserUsecase
}

func NewUserHandler(rg *gin.RouterGroup, us domain.UserUsecase) {
	handler := &UserHandler{
		UserUsecase: us,
	}

	usersGroup := rg.Group("/users")
	{
		usersGroup.GET("", handler.Fetch)
		usersGroup.GET("/:id", handler.GetByID)
		usersGroup.POST("", handler.Create)
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if err := h.UserUsecase.Create(c.Request.Context(), &user); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, "Email already exists", err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to create user", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "User created successfully", user)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "The provided ID is not a valid UUID")
		return
	}

	user, err := h.UserUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User fetched successfully", user)
}

func (h *UserHandler) Fetch(c *gin.Context) {
	page := 1
	perPage := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 && v <= 100 {
			perPage = v
		}
	}

	params := domain.PaginationParams{Page: page, PerPage: perPage}

	result, err := h.UserUsecase.Fetch(c.Request.Context(), params)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch users", err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Users fetched successfully", result.Items, map[string]int{
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_items": result.TotalItems,
		"total_pages": result.TotalPages,
	})
}
