package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

type TransactionHandler struct {
	transactionUsecase domain.TransactionUsecase
}

func NewTransactionHandler(rg *gin.RouterGroup, uc domain.TransactionUsecase) {
	handler := &TransactionHandler{transactionUsecase: uc}
	rg.POST("/transactions", handler.Create)
	rg.GET("/transactions", handler.Fetch)
	rg.PUT("/transactions/:id", handler.Update)
	rg.DELETE("/transactions/:id", handler.Delete)
}

// Create godoc
// @Summary      Create a new transaction
// @Description  Create a new expense, income, or transfer transaction
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        request body domain.CreateTransactionRequest true "Transaction Data"
// @Success      201  {object}  response.Response{data=domain.Transaction}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	result, err := h.transactionUsecase.Create(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Transaction created successfully", result)
}

// Fetch godoc
// @Summary      Get all transactions
// @Description  Fetch a paginated list of transactions with optional filters
// @Tags         transactions
// @Produce      json
// @Param        page               query  int     false  "Page number"
// @Param        per_page           query  int     false  "Items per page"
// @Param        account_id         query  string  false  "Filter by account ID"
// @Param        transaction_type   query  string  false  "Filter by type (EXPENSE/INCOME/TRANSFER)"
// @Success      200                {object}  response.Response{data=[]domain.Transaction}
// @Failure      500                {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/transactions [get]
func (h *TransactionHandler) Fetch(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	params := domain.TransactionQueryParams{
		UserID:  userID.(uuid.UUID),
		Page:    page,
		PerPage: perPage,
	}

	if aid := c.Query("account_id"); aid != "" {
		if id, err := uuid.Parse(aid); err == nil {
			params.AccountID = &id
		}
	}
	if tt := c.Query("transaction_type"); tt != "" {
		txType := domain.TransactionType(tt)
		params.TransactionType = &txType
	}

	result, err := h.transactionUsecase.Fetch(c.Request.Context(), params)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Transactions fetched successfully", result.Items, map[string]int{
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_items": result.TotalItems,
		"total_pages": result.TotalPages,
	})
}

// Update godoc
// @Summary      Update a transaction
// @Description  Update details of a specific transaction by ID
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        id      path      string                           true  "Transaction ID"
// @Param        request body      domain.UpdateTransactionRequest  true  "Updated Transaction Data"
// @Success      200     {object}  response.Response{data=domain.Transaction}
// @Failure      400     {object}  response.Response
// @Failure      500     {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/transactions/{id} [put]
func (h *TransactionHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", "The provided ID is not a valid UUID")
		return
	}

	var req domain.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	result, err := h.transactionUsecase.Update(c.Request.Context(), userID.(uuid.UUID), id, &req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Transaction updated successfully", result)
}

// Delete godoc
// @Summary      Delete a transaction
// @Description  Soft delete a specific transaction by ID
// @Tags         transactions
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", "The provided ID is not a valid UUID")
		return
	}

	if err := h.transactionUsecase.SoftDelete(c.Request.Context(), userID.(uuid.UUID), id); err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Transaction deleted successfully", nil)
}
