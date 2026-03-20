package http_test

import (
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

func setupDashboardRouter(mockUC *mocks.MockDashboardUsecase) (*gin.Engine, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	userID := uuid.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	v1 := r.Group("/api/v1")
	handler.NewDashboardHandler(v1, mockUC)
	return r, userID
}

func TestDashboardHandler_GetSummary_Success(t *testing.T) {
	mockUC := new(mocks.MockDashboardUsecase)
	r, userID := setupDashboardRouter(mockUC)

	expectedSummary := &domain.DashboardSummary{
		TotalBalance: 1500000,
		RecentTransactions: []domain.Transaction{
			{
				ID:              uuid.New(),
				UserID:          userID,
				AccountID:       uuid.New(),
				TransactionType: domain.TxTypeExpense,
				Amount:          50000,
				Note:            "Makan siang",
				TransactionDate: time.Now(),
				CreatedAt:       time.Now(),
			},
		},
		NeedsPaydaySetup: false,
	}
	mockUC.On("GetSummary", mock.Anything, userID).Return(expectedSummary, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Dashboard summary fetched successfully", res.Message)
	mockUC.AssertExpectations(t)
}
