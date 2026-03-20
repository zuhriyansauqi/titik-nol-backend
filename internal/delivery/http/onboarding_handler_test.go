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

func setupOnboardingRouter(mockUC *mocks.MockOnboardingUsecase) (*gin.Engine, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := uuid.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	v1 := r.Group("/api/v1")
	handler.NewOnboardingHandler(v1, mockUC)
	return r, userID
}

func TestOnboardingHandler_SetupAccounts_Success(t *testing.T) {
	mockUC := new(mocks.MockOnboardingUsecase)
	r, userID := setupOnboardingRouter(mockUC)

	reqBody := domain.SetupAccountsRequest{
		Accounts: []domain.SetupAccountItem{
			{Name: "Cash", Type: domain.AccountTypeCash, InitialBalance: 500000},
			{Name: "Bank BCA", Type: domain.AccountTypeBank, InitialBalance: 0},
		},
	}
	body, _ := json.Marshal(reqBody)

	expectedResp := &domain.SetupAccountsResponse{
		Accounts: []domain.Account{
			{ID: uuid.New(), UserID: userID, Name: "Cash", Type: domain.AccountTypeCash, Balance: 500000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New(), UserID: userID, Name: "Bank BCA", Type: domain.AccountTypeBank, Balance: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		Transactions: []domain.Transaction{
			{ID: uuid.New(), UserID: userID, AccountID: uuid.New(), TransactionType: domain.TxTypeAdjustment, Amount: 500000, Note: "Saldo awal (onboarding)", TransactionDate: time.Now(), CreatedAt: time.Now()},
		},
	}
	mockUC.On("SetupAccounts", mock.Anything, userID, mock.AnythingOfType("*domain.SetupAccountsRequest")).Return(expectedResp, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/onboarding/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Accounts setup successfully", res.Message)
	mockUC.AssertExpectations(t)
}

func TestOnboardingHandler_SetupAccounts_InvalidBody(t *testing.T) {
	mockUC := new(mocks.MockOnboardingUsecase)
	r, _ := setupOnboardingRouter(mockUC)

	// Missing required fields in account items
	body := []byte(`{"accounts": [{"initial_balance": 100}]}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/onboarding/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	mockUC.AssertNotCalled(t, "SetupAccounts")
}

func TestOnboardingHandler_SetupAccounts_EmptyList(t *testing.T) {
	mockUC := new(mocks.MockOnboardingUsecase)
	r, _ := setupOnboardingRouter(mockUC)

	// Empty accounts array — fails binding validation (min=1)
	body := []byte(`{"accounts": []}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/onboarding/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	mockUC.AssertNotCalled(t, "SetupAccounts")
}
