package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

type AccountHandler struct {
	accountUsecase domain.AccountUsecase
}

func NewAccountHandler(rg *gin.RouterGroup, uc domain.AccountUsecase) {
	handler := &AccountHandler{accountUsecase: uc}
	rg.GET("/accounts", handler.Fetch)
	rg.POST("/accounts", handler.Create)
	rg.PUT("/accounts/:id", handler.Update)
	rg.DELETE("/accounts/:id", handler.Delete)
}

func (h *AccountHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	account, err := h.accountUsecase.Create(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Account created successfully", account)
}

func (h *AccountHandler) Fetch(c *gin.Context) {
	userID, _ := c.Get("user_id")

	accounts, err := h.accountUsecase.FetchByUserID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Accounts fetched successfully", accounts)
}

func (h *AccountHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID", "The provided ID is not a valid UUID")
		return
	}

	var req domain.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	account, err := h.accountUsecase.Update(c.Request.Context(), userID.(uuid.UUID), id, &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Account updated successfully", account)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID", "The provided ID is not a valid UUID")
		return
	}

	if err := h.accountUsecase.SoftDelete(c.Request.Context(), userID.(uuid.UUID), id); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Account deleted successfully", nil)
}
