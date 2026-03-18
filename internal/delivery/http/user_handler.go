package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.UserUsecase.Create(c.Request.Context(), &user); err != nil {
		if err == domain.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.UserUsecase.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Fetch(c *gin.Context) {
	users, err := h.UserUsecase.Fetch(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
