package usecase_test

import (
	"context"
	"errors"
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

func TestTransactionUsecase_Create_Income(t *testing.T) {
	db, smock := newTestDB(t)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	txTxRepo := new(mocks.MockTransactionRepository)
	txAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)

	userID := uuid.New()
	accountID := uuid.New()
	account := &domain.Account{
		ID:      accountID,
		UserID:  userID,
		Name:    "Bank BCA",
		Type:    domain.AccountTypeBank,
		Balance: 100000,
	}

	req := &domain.CreateTransactionRequest{
		AccountID:       accountID,
		TransactionType: domain.TxTypeIncome,
		Amount:          50000,
		Note:            "Salary",
		TransactionDate: time.Now(),
	}

	txAccRepo.On("GetByID", mock.Anything, accountID, userID).Return(account, nil)
	txTxRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)
	txAccRepo.On("UpdateBalance", mock.Anything, accountID, int64(50000)).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewTransactionUsecase(mockTxRepo, mockAccRepo, mockCatRepo, db)
	result, err := uc.Create(context.Background(), userID, req)

	require.NoError(t, err)
	assert.Equal(t, domain.TxTypeIncome, result.Transaction.TransactionType)
	assert.Equal(t, int64(50000), result.Transaction.Amount)
	assert.Equal(t, int64(150000), result.AccountBalance)
	txAccRepo.AssertCalled(t, "UpdateBalance", mock.Anything, accountID, int64(50000))
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestTransactionUsecase_Create_Expense(t *testing.T) {
	db, smock := newTestDB(t)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	txTxRepo := new(mocks.MockTransactionRepository)
	txAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)

	userID := uuid.New()
	accountID := uuid.New()
	account := &domain.Account{
		ID:      accountID,
		UserID:  userID,
		Name:    "Cash",
		Type:    domain.AccountTypeCash,
		Balance: 200000,
	}

	req := &domain.CreateTransactionRequest{
		AccountID:       accountID,
		TransactionType: domain.TxTypeExpense,
		Amount:          75000,
		Note:            "Groceries",
		TransactionDate: time.Now(),
	}

	txAccRepo.On("GetByID", mock.Anything, accountID, userID).Return(account, nil)
	txTxRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)
	txAccRepo.On("UpdateBalance", mock.Anything, accountID, int64(-75000)).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewTransactionUsecase(mockTxRepo, mockAccRepo, mockCatRepo, db)
	result, err := uc.Create(context.Background(), userID, req)

	require.NoError(t, err)
	assert.Equal(t, domain.TxTypeExpense, result.Transaction.TransactionType)
	assert.Equal(t, int64(75000), result.Transaction.Amount)
	assert.Equal(t, int64(125000), result.AccountBalance)
	txAccRepo.AssertCalled(t, "UpdateBalance", mock.Anything, accountID, int64(-75000))
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestTransactionUsecase_Create_AccountNotFound(t *testing.T) {
	db, smock := newTestDB(t)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	txTxRepo := new(mocks.MockTransactionRepository)
	txAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)

	userID := uuid.New()
	accountID := uuid.New()

	req := &domain.CreateTransactionRequest{
		AccountID:       accountID,
		TransactionType: domain.TxTypeIncome,
		Amount:          50000,
		TransactionDate: time.Now(),
	}

	txAccRepo.On("GetByID", mock.Anything, accountID, userID).Return(nil, errors.New("not found"))

	smock.ExpectBegin()
	smock.ExpectRollback()

	uc := usecase.NewTransactionUsecase(mockTxRepo, mockAccRepo, mockCatRepo, db)
	result, err := uc.Create(context.Background(), userID, req)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	txTxRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestTransactionUsecase_Create_WithCategory(t *testing.T) {
	db, smock := newTestDB(t)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	txTxRepo := new(mocks.MockTransactionRepository)
	txAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)

	userID := uuid.New()
	accountID := uuid.New()
	categoryID := uuid.New()
	account := &domain.Account{
		ID:      accountID,
		UserID:  userID,
		Name:    "Bank BCA",
		Type:    domain.AccountTypeBank,
		Balance: 300000,
	}
	category := &domain.Category{
		ID:     categoryID,
		UserID: userID,
		Name:   "Food",
		Type:   domain.CategoryTypeExpense,
	}

	req := &domain.CreateTransactionRequest{
		AccountID:       accountID,
		CategoryID:      &categoryID,
		TransactionType: domain.TxTypeExpense,
		Amount:          25000,
		Note:            "Lunch",
		TransactionDate: time.Now(),
	}

	txAccRepo.On("GetByID", mock.Anything, accountID, userID).Return(account, nil)
	mockCatRepo.On("GetByID", mock.Anything, categoryID, userID).Return(category, nil)
	txTxRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)
	txAccRepo.On("UpdateBalance", mock.Anything, accountID, int64(-25000)).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewTransactionUsecase(mockTxRepo, mockAccRepo, mockCatRepo, db)
	result, err := uc.Create(context.Background(), userID, req)

	require.NoError(t, err)
	assert.Equal(t, &categoryID, result.Transaction.CategoryID)
	assert.Equal(t, int64(275000), result.AccountBalance)
	mockCatRepo.AssertCalled(t, "GetByID", mock.Anything, categoryID, userID)
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestTransactionUsecase_Update_AmountChanged(t *testing.T) {
	db, smock := newTestDB(t)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	txTxRepo := new(mocks.MockTransactionRepository)
	txAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)

	userID := uuid.New()
	txID := uuid.New()
	accountID := uuid.New()
	txDate := time.Now()

	existingTx := &domain.Transaction{
		ID:              txID,
		UserID:          userID,
		AccountID:       accountID,
		TransactionType: domain.TxTypeExpense,
		Amount:          50000,
		Note:            "Old note",
		TransactionDate: txDate,
	}

	// Old delta: -50000, New delta: -80000, adjustment: -30000
	updatedAccount := &domain.Account{
		ID:      accountID,
		UserID:  userID,
		Name:    "Cash",
		Balance: 70000,
	}

	req := &domain.UpdateTransactionRequest{
		Amount:          80000,
		Note:            "Updated note",
		TransactionDate: txDate,
	}

	txTxRepo.On("GetByID", mock.Anything, txID, userID).Return(existingTx, nil)
	txTxRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)
	txAccRepo.On("UpdateBalance", mock.Anything, accountID, int64(-30000)).Return(nil)
	txAccRepo.On("GetByID", mock.Anything, accountID, userID).Return(updatedAccount, nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewTransactionUsecase(mockTxRepo, mockAccRepo, mockCatRepo, db)
	result, err := uc.Update(context.Background(), userID, txID, req)

	require.NoError(t, err)
	assert.Equal(t, int64(80000), result.Transaction.Amount)
	assert.Equal(t, "Updated note", result.Transaction.Note)
	assert.Equal(t, int64(70000), result.AccountBalance)
	txAccRepo.AssertCalled(t, "UpdateBalance", mock.Anything, accountID, int64(-30000))
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestTransactionUsecase_SoftDelete_Reversal(t *testing.T) {
	db, smock := newTestDB(t)
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	txTxRepo := new(mocks.MockTransactionRepository)
	txAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)

	userID := uuid.New()
	txID := uuid.New()
	accountID := uuid.New()

	// Deleting an EXPENSE of 60000 should reverse with +60000
	existingTx := &domain.Transaction{
		ID:              txID,
		UserID:          userID,
		AccountID:       accountID,
		TransactionType: domain.TxTypeExpense,
		Amount:          60000,
		TransactionDate: time.Now(),
	}

	txTxRepo.On("GetByID", mock.Anything, txID, userID).Return(existingTx, nil)
	txTxRepo.On("SoftDelete", mock.Anything, txID, userID).Return(nil)
	txAccRepo.On("UpdateBalance", mock.Anything, accountID, int64(60000)).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewTransactionUsecase(mockTxRepo, mockAccRepo, mockCatRepo, db)
	err := uc.SoftDelete(context.Background(), userID, txID)

	assert.NoError(t, err)
	txAccRepo.AssertCalled(t, "UpdateBalance", mock.Anything, accountID, int64(60000))
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestTransactionUsecase_Fetch_WithPagination(t *testing.T) {
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)

	userID := uuid.New()
	params := domain.TransactionQueryParams{
		UserID:  userID,
		Page:    2,
		PerPage: 10,
	}

	transactions := []domain.Transaction{
		{ID: uuid.New(), UserID: userID, Amount: 10000, TransactionType: domain.TxTypeIncome},
		{ID: uuid.New(), UserID: userID, Amount: 20000, TransactionType: domain.TxTypeExpense},
	}

	mockTxRepo.On("Fetch", mock.Anything, params).Return(transactions, 25, nil)

	uc := usecase.NewTransactionUsecase(mockTxRepo, mockAccRepo, mockCatRepo, nil)
	result, err := uc.Fetch(context.Background(), params)

	require.NoError(t, err)
	assert.Equal(t, 25, result.TotalItems)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.Equal(t, 3, result.TotalPages)
	items := result.Items.([]domain.Transaction)
	assert.Len(t, items, 2)
	mockTxRepo.AssertExpectations(t)
}
