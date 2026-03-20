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

func setupCategoryRouter(mockUC *mocks.MockCategoryUsecase) (*gin.Engine, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := uuid.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	v1 := r.Group("/api/v1")
	handler.NewCategoryHandler(v1, mockUC)
	return r, userID
}

func TestCategoryHandler_BulkCreate_Success(t *testing.T) {
	mockUC := new(mocks.MockCategoryUsecase)
	r, userID := setupCategoryRouter(mockUC)

	reqBody := domain.BulkCreateCategoryRequest{
		Categories: []domain.BulkCreateCategoryItem{
			{Name: "Gaji", Type: domain.CategoryTypeIncome, Icon: "💰"},
			{Name: "Makan", Type: domain.CategoryTypeExpense, Icon: "🍔"},
		},
	}
	body, _ := json.Marshal(reqBody)

	expectedCategories := []domain.Category{
		{ID: uuid.New(), UserID: userID, Name: "Gaji", Type: domain.CategoryTypeIncome, Icon: "💰", CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: userID, Name: "Makan", Type: domain.CategoryTypeExpense, Icon: "🍔", CreatedAt: time.Now()},
	}
	mockUC.On("BulkCreate", mock.Anything, userID, mock.AnythingOfType("*domain.BulkCreateCategoryRequest")).Return(expectedCategories, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Categories created successfully", res.Message)
	mockUC.AssertExpectations(t)
}

func TestCategoryHandler_BulkCreate_InvalidBody(t *testing.T) {
	mockUC := new(mocks.MockCategoryUsecase)
	r, _ := setupCategoryRouter(mockUC)

	// Missing required fields (name and type)
	body := []byte(`{"categories": [{"icon": "x"}]}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	mockUC.AssertNotCalled(t, "BulkCreate")
}

func TestCategoryHandler_Fetch_Success(t *testing.T) {
	mockUC := new(mocks.MockCategoryUsecase)
	r, userID := setupCategoryRouter(mockUC)

	categories := []domain.Category{
		{ID: uuid.New(), UserID: userID, Name: "Gaji", Type: domain.CategoryTypeIncome, CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: userID, Name: "Makan", Type: domain.CategoryTypeExpense, CreatedAt: time.Now()},
	}
	mockUC.On("FetchByUserID", mock.Anything, userID, (*domain.CategoryType)(nil)).Return(categories, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Categories fetched successfully", res.Message)
	mockUC.AssertExpectations(t)
}

func TestCategoryHandler_Fetch_WithTypeFilter(t *testing.T) {
	mockUC := new(mocks.MockCategoryUsecase)
	r, userID := setupCategoryRouter(mockUC)

	expenseType := domain.CategoryTypeExpense
	categories := []domain.Category{
		{ID: uuid.New(), UserID: userID, Name: "Makan", Type: domain.CategoryTypeExpense, CreatedAt: time.Now()},
	}
	mockUC.On("FetchByUserID", mock.Anything, userID, &expenseType).Return(categories, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories?type=EXPENSE", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Categories fetched successfully", res.Message)
	mockUC.AssertExpectations(t)
}
