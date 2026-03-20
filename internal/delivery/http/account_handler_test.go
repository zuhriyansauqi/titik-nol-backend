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

func setupAccountRouter(mockUC *mocks.MockAccountUsecase) (*gin.Engine, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := uuid.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	v1 := r.Group("/api/v1")
	handler.NewAccountHandler(v1, mockUC)
	return r, userID
}

func TestAccountHandler_Create_Success(t *testing.T) {
	mockUC := new(mocks.MockAccountUsecase)
	r, userID := setupAccountRouter(mockUC)

	reqBody := domain.CreateAccountRequest{
		Name:           "Bank BCA",
		Type:           domain.AccountTypeCash,
		InitialBalance: 100000,
	}
	body, _ := json.Marshal(reqBody)

	expectedAccount := &domain.Account{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "Bank BCA",
		Type:      domain.AccountTypeCash,
		Balance:   100000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockUC.On("Create", mock.Anything, userID, mock.AnythingOfType("*domain.CreateAccountRequest")).Return(expectedAccount, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Account created successfully", res.Message)
	mockUC.AssertExpectations(t)
}

func TestAccountHandler_Create_InvalidBody(t *testing.T) {
	mockUC := new(mocks.MockAccountUsecase)
	r, _ := setupAccountRouter(mockUC)

	// Missing required fields (name and type)
	body := []byte(`{"initial_balance": 100}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	mockUC.AssertNotCalled(t, "Create")
}

func TestAccountHandler_Fetch_Success(t *testing.T) {
	mockUC := new(mocks.MockAccountUsecase)
	r, userID := setupAccountRouter(mockUC)

	accounts := []domain.Account{
		{ID: uuid.New(), UserID: userID, Name: "Cash", Type: domain.AccountTypeCash, Balance: 50000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), UserID: userID, Name: "Bank", Type: domain.AccountTypeBank, Balance: 200000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	mockUC.On("FetchByUserID", mock.Anything, userID).Return(accounts, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Accounts fetched successfully", res.Message)
	mockUC.AssertExpectations(t)
}

func TestAccountHandler_Update_NotFound(t *testing.T) {
	mockUC := new(mocks.MockAccountUsecase)
	r, userID := setupAccountRouter(mockUC)

	accountID := uuid.New()
	mockUC.On("Update", mock.Anything, userID, accountID, mock.AnythingOfType("*domain.UpdateAccountRequest")).Return(nil, domain.ErrAccountNotFound)

	reqBody := domain.UpdateAccountRequest{Name: "Updated Name"}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/accounts/"+accountID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	mockUC.AssertExpectations(t)
}

func TestAccountHandler_Delete_Success(t *testing.T) {
	mockUC := new(mocks.MockAccountUsecase)
	r, userID := setupAccountRouter(mockUC)

	accountID := uuid.New()
	mockUC.On("SoftDelete", mock.Anything, userID, accountID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/accounts/"+accountID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Account deleted successfully", res.Message)
	mockUC.AssertExpectations(t)
}
