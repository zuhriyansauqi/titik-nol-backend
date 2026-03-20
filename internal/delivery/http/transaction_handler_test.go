package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	handler "github.com/mzhryns/titik-nol-backend/internal/delivery/http"
)

func setupTransactionRouter(mockUC *mocks.MockTransactionUsecase) (*gin.Engine, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := uuid.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	v1 := r.Group("/api/v1")
	handler.NewTransactionHandler(v1, mockUC)
	return r, userID
}

func TestTransactionHandler_Create_Success(t *testing.T) {
	mockUC := new(mocks.MockTransactionUsecase)
	r, userID := setupTransactionRouter(mockUC)

	accountID := uuid.New()
	txDate := time.Now().Truncate(time.Second)

	reqBody := domain.CreateTransactionRequest{
		AccountID:       accountID,
		TransactionType: domain.TxTypeIncome,
		Amount:          50000,
		Note:            "Salary",
		TransactionDate: txDate,
	}
	body, _ := json.Marshal(reqBody)

	expectedResult := &domain.CreateTransactionResponse{
		Transaction: domain.Transaction{
			ID:              uuid.New(),
			UserID:          userID,
			AccountID:       accountID,
			TransactionType: domain.TxTypeIncome,
			Amount:          50000,
			Note:            "Salary",
			TransactionDate: txDate,
			CreatedAt:       time.Now(),
		},
		AccountBalance: 150000,
	}
	mockUC.On("Create", mock.Anything, userID, mock.AnythingOfType("*domain.CreateTransactionRequest")).Return(expectedResult, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Transaction created successfully", res.Message)
	mockUC.AssertExpectations(t)
}

func TestTransactionHandler_Create_InvalidBody(t *testing.T) {
	mockUC := new(mocks.MockTransactionUsecase)
	r, _ := setupTransactionRouter(mockUC)

	// Missing required fields (amount, transaction_type, account_id, transaction_date)
	body := []byte(`{"note": "incomplete"}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	mockUC.AssertNotCalled(t, "Create")
}

func TestTransactionHandler_Fetch_WithPagination(t *testing.T) {
	mockUC := new(mocks.MockTransactionUsecase)
	r, userID := setupTransactionRouter(mockUC)

	accountID := uuid.New()
	txType := domain.TxTypeExpense

	expectedParams := domain.TransactionQueryParams{
		UserID:          userID,
		AccountID:       &accountID,
		TransactionType: &txType,
		Page:            2,
		PerPage:         10,
	}

	expectedResult := &domain.PaginatedResult{
		Items: []domain.Transaction{
			{ID: uuid.New(), UserID: userID, AccountID: accountID, TransactionType: domain.TxTypeExpense, Amount: 25000},
		},
		TotalItems: 15,
		Page:       2,
		PerPage:    10,
		TotalPages: 2,
	}
	mockUC.On("Fetch", mock.Anything, expectedParams).Return(expectedResult, nil)

	w := httptest.NewRecorder()
	url := "/api/v1/transactions?page=2&per_page=10&account_id=" + accountID.String() + "&transaction_type=EXPENSE"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Transactions fetched successfully", res.Message)
	assert.NotNil(t, res.Meta)
	mockUC.AssertExpectations(t)
}

func TestTransactionHandler_Update_Success(t *testing.T) {
	mockUC := new(mocks.MockTransactionUsecase)
	r, userID := setupTransactionRouter(mockUC)

	txID := uuid.New()
	accountID := uuid.New()
	txDate := time.Now().Truncate(time.Second)

	reqBody := domain.UpdateTransactionRequest{
		Amount:          75000,
		Note:            "Updated note",
		TransactionDate: txDate,
	}
	body, _ := json.Marshal(reqBody)

	expectedResult := &domain.UpdateTransactionResponse{
		Transaction: domain.Transaction{
			ID:              txID,
			UserID:          userID,
			AccountID:       accountID,
			TransactionType: domain.TxTypeExpense,
			Amount:          75000,
			Note:            "Updated note",
			TransactionDate: txDate,
		},
		AccountBalance: 125000,
	}
	mockUC.On("Update", mock.Anything, userID, txID, mock.AnythingOfType("*domain.UpdateTransactionRequest")).Return(expectedResult, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/transactions/"+txID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Transaction updated successfully", res.Message)
	mockUC.AssertExpectations(t)
}

func TestTransactionHandler_Delete_NotFound(t *testing.T) {
	mockUC := new(mocks.MockTransactionUsecase)
	r, userID := setupTransactionRouter(mockUC)

	txID := uuid.New()
	mockUC.On("SoftDelete", mock.Anything, userID, txID).Return(domain.ErrTransactionNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/transactions/"+txID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	mockUC.AssertExpectations(t)
}
