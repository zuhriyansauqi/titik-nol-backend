package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDashboardUsecase_GetSummary_WithCategories(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	userID := uuid.New()

	mockAccRepo.On("FetchByUserID", mock.Anything, userID).Return([]domain.Account{
		{ID: uuid.New(), UserID: userID, Name: "Cash", Type: domain.AccountTypeCash, Balance: 100000},
	}, nil)

	mockTxRepo.On("FetchRecent", mock.Anything, userID, 5).Return([]domain.Transaction{
		{ID: uuid.New(), UserID: userID, Amount: 50000, TransactionType: domain.TxTypeIncome, TransactionDate: time.Now()},
	}, nil)

	mockCatRepo.On("CountByUserID", mock.Anything, userID).Return(int64(3), nil)

	uc := usecase.NewDashboardUsecase(mockAccRepo, mockTxRepo, mockCatRepo)
	summary, err := uc.GetSummary(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, int64(100000), summary.TotalBalance)
	assert.Len(t, summary.RecentTransactions, 1)
	assert.False(t, summary.NeedsPaydaySetup)

	mockAccRepo.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestDashboardUsecase_GetSummary_NoCategories(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	userID := uuid.New()

	mockAccRepo.On("FetchByUserID", mock.Anything, userID).Return([]domain.Account{
		{ID: uuid.New(), UserID: userID, Name: "Bank", Type: domain.AccountTypeBank, Balance: 200000},
	}, nil)

	mockTxRepo.On("FetchRecent", mock.Anything, userID, 5).Return([]domain.Transaction{}, nil)

	mockCatRepo.On("CountByUserID", mock.Anything, userID).Return(int64(0), nil)

	uc := usecase.NewDashboardUsecase(mockAccRepo, mockTxRepo, mockCatRepo)
	summary, err := uc.GetSummary(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, int64(200000), summary.TotalBalance)
	assert.Empty(t, summary.RecentTransactions)
	assert.True(t, summary.NeedsPaydaySetup)

	mockAccRepo.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestDashboardUsecase_GetSummary_TotalBalance(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	userID := uuid.New()

	mockAccRepo.On("FetchByUserID", mock.Anything, userID).Return([]domain.Account{
		{ID: uuid.New(), UserID: userID, Name: "Cash", Type: domain.AccountTypeCash, Balance: 150000},
		{ID: uuid.New(), UserID: userID, Name: "Bank BCA", Type: domain.AccountTypeBank, Balance: 500000},
		{ID: uuid.New(), UserID: userID, Name: "GoPay", Type: domain.AccountTypeEWallet, Balance: 75000},
	}, nil)

	mockTxRepo.On("FetchRecent", mock.Anything, userID, 5).Return([]domain.Transaction{}, nil)

	mockCatRepo.On("CountByUserID", mock.Anything, userID).Return(int64(1), nil)

	uc := usecase.NewDashboardUsecase(mockAccRepo, mockTxRepo, mockCatRepo)
	summary, err := uc.GetSummary(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, int64(725000), summary.TotalBalance) // 150000 + 500000 + 75000
	assert.False(t, summary.NeedsPaydaySetup)

	mockAccRepo.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}
