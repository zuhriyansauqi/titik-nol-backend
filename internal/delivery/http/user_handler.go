package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

type UserHandler struct {
	UserUsecase domain.UserUsecase
}

func NewUserHandler(r *gin.Engine, us domain.UserUsecase) {
	handler := &UserHandler{
		UserUsecase: us,
	}

	v1 := r.Group("/api/v1")
	{
		v1.GET("/users", handler.Fetch)
		v1.GET("/users/:id", handler.GetByID)
		v1.POST("/users", handler.Create)
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if err := h.UserUsecase.Create(c.Request.Context(), &user); err != nil {
		if err == domain.ErrEmailAlreadyExists {
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
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "The provided ID is not a valid integer")
		return
	}

	user, err := h.UserUsecase.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, http.StatusOK, "User fetched successfully", user)
}

func (h *UserHandler) Fetch(c *gin.Context) {
	users, err := h.UserUsecase.Fetch(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to fetch users", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Users fetched successfully", users)
}
