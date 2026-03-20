package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOnboardingUsecase_SetupAccounts_Success(t *testing.T) {
	db, smock := newTestDB(t)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	txAccRepo := new(mocks.MockAccountRepository)
	txTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)

	userID := uuid.New()
	req := &domain.SetupAccountsRequest{
		Accounts: []domain.SetupAccountItem{
			{Name: "Cash", Type: domain.AccountTypeCash, InitialBalance: 100000},
			{Name: "Bank BCA", Type: domain.AccountTypeBank, InitialBalance: 500000},
			{Name: "GoPay", Type: domain.AccountTypeEWallet, InitialBalance: 0},
		},
	}

	txAccRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)
	txTxRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewOnboardingUsecase(mockAccRepo, mockTxRepo, db)
	resp, err := uc.SetupAccounts(context.Background(), userID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Accounts, 3)
	assert.Len(t, resp.Transactions, 2) // only 2 accounts have InitialBalance > 0

	// Verify account names
	assert.Equal(t, "Cash", resp.Accounts[0].Name)
	assert.Equal(t, "Bank BCA", resp.Accounts[1].Name)
	assert.Equal(t, "GoPay", resp.Accounts[2].Name)

	// Verify ADJUSTMENT transactions have correct amounts
	assert.Equal(t, int64(100000), resp.Transactions[0].Amount)
	assert.Equal(t, domain.TxTypeAdjustment, resp.Transactions[0].TransactionType)
	assert.Equal(t, int64(500000), resp.Transactions[1].Amount)
	assert.Equal(t, domain.TxTypeAdjustment, resp.Transactions[1].TransactionType)

	txAccRepo.AssertNumberOfCalls(t, "Create", 3)
	txTxRepo.AssertNumberOfCalls(t, "Create", 2)
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestOnboardingUsecase_SetupAccounts_ZeroBalance(t *testing.T) {
	db, smock := newTestDB(t)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	txAccRepo := new(mocks.MockAccountRepository)
	txTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)

	userID := uuid.New()
	req := &domain.SetupAccountsRequest{
		Accounts: []domain.SetupAccountItem{
			{Name: "Cash", Type: domain.AccountTypeCash, InitialBalance: 0},
		},
	}

	txAccRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewOnboardingUsecase(mockAccRepo, mockTxRepo, db)
	resp, err := uc.SetupAccounts(context.Background(), userID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Accounts, 1)
	assert.Empty(t, resp.Transactions) // no ADJUSTMENT for zero balance

	txAccRepo.AssertNumberOfCalls(t, "Create", 1)
	txTxRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestOnboardingUsecase_SetupAccounts_EmptyList(t *testing.T) {
	db, _ := newTestDB(t)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	userID := uuid.New()
	req := &domain.SetupAccountsRequest{
		Accounts: []domain.SetupAccountItem{},
	}

	uc := usecase.NewOnboardingUsecase(mockAccRepo, mockTxRepo, db)
	resp, err := uc.SetupAccounts(context.Background(), userID, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrEmptyBulkRequest)

	mockAccRepo.AssertNotCalled(t, "WithTx", mock.Anything)
}

func TestOnboardingUsecase_SetupAccounts_InvalidItem(t *testing.T) {
	db, _ := newTestDB(t)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	userID := uuid.New()
	req := &domain.SetupAccountsRequest{
		Accounts: []domain.SetupAccountItem{
			{Name: "Cash", Type: domain.AccountTypeCash, InitialBalance: 100000},
			{Name: "", Type: domain.AccountTypeBank, InitialBalance: 500000}, // invalid: empty name
		},
	}

	uc := usecase.NewOnboardingUsecase(mockAccRepo, mockTxRepo, db)
	resp, err := uc.SetupAccounts(context.Background(), userID, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrValidationFailed)
	assert.Contains(t, err.Error(), "account[1]") // error should reference index 1

	mockAccRepo.AssertNotCalled(t, "WithTx", mock.Anything)
}
