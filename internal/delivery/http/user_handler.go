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

func NewUserHandler(rg *gin.RouterGroup, us domain.UserUsecase, adminMiddleware gin.HandlerFunc) {
	handler := &UserHandler{
		UserUsecase: us,
	}

	// Self-service profile routes (any authenticated user)
	rg.GET("/users/me", handler.GetProfile)
	rg.PUT("/users/me", handler.UpdateProfile)

	// Admin-only routes
	usersGroup := rg.Group("/users")
	usersGroup.Use(adminMiddleware)
	{
		usersGroup.GET("", handler.Fetch)
		usersGroup.GET("/:id", handler.GetByID)
		usersGroup.POST("", handler.Create)
	}
}

// GetProfile godoc
// @Summary      Get own profile
// @Description  Get the authenticated user's profile
// @Tags         users
// @Produce      json
// @Success      200  {object}  response.Response{data=domain.User}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found in context", nil)
		return
	}

	user, err := h.UserUsecase.GetByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User profile fetched successfully", user)
}

// UpdateProfile godoc
// @Summary      Update own profile
// @Description  Update the authenticated user's name or avatar
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body domain.UpdateProfileRequest true "Update Profile Request"
// @Success      200  {object}  response.Response{data=domain.User}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found in context", nil)
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	user, err := h.UserUsecase.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User profile updated successfully", user)
}

// Create godoc
// @Summary      Create a new user
// @Description  Create a new user with the provided details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body domain.User true "User Data"
// @Success      201  {object}  response.Response{data=domain.User}
// @Failure      400  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/users [post]
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

// GetByID godoc
// @Summary      Get user by ID
// @Description  Fetch a single user by their UUID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.Response{data=domain.User}
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [get]
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

// Fetch godoc
// @Summary      Get all users
// @Description  Fetch a paginated list of users
// @Tags         users
// @Produce      json
// @Param        page      query  int  false  "Page number"
// @Param        per_page  query  int  false  "Items per page"
// @Success      200       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/users [get]
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
