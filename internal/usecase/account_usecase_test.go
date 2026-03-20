package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newTestDB creates a *gorm.DB backed by sqlmock for testing GORM transactions.
func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, smock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	dialector := gorm_postgres.New(gorm_postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)
	return db, smock
}

func TestAccountUsecase_Create_Success(t *testing.T) {
	db, smock := newTestDB(t)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	txAccRepo := new(mocks.MockAccountRepository)
	txTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)

	userID := uuid.New()
	req := &domain.CreateAccountRequest{
		Name:           "Bank BCA",
		Type:           domain.AccountTypeBank,
		InitialBalance: 500000,
	}

	txAccRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)
	txTxRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	// sqlmock expects BEGIN and COMMIT for the GORM transaction
	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewAccountUsecase(mockAccRepo, mockTxRepo, db)
	account, err := uc.Create(context.Background(), userID, req)

	require.NoError(t, err)
	assert.Equal(t, "Bank BCA", account.Name)
	assert.Equal(t, domain.AccountTypeBank, account.Type)
	assert.Equal(t, int64(500000), account.Balance)
	assert.Equal(t, userID, account.UserID)

	txAccRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*domain.Account"))
	txTxRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*domain.Transaction"))
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestAccountUsecase_Create_ZeroBalance(t *testing.T) {
	db, smock := newTestDB(t)
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	txAccRepo := new(mocks.MockAccountRepository)
	txTxRepo := new(mocks.MockTransactionRepository)
	mockAccRepo.On("WithTx", mock.Anything).Return(txAccRepo)
	mockTxRepo.On("WithTx", mock.Anything).Return(txTxRepo)

	userID := uuid.New()
	req := &domain.CreateAccountRequest{
		Name:           "Cash",
		Type:           domain.AccountTypeCash,
		InitialBalance: 0,
	}

	txAccRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewAccountUsecase(mockAccRepo, mockTxRepo, db)
	account, err := uc.Create(context.Background(), userID, req)

	require.NoError(t, err)
	assert.Equal(t, "Cash", account.Name)
	assert.Equal(t, int64(0), account.Balance)

	// ADJUSTMENT transaction should NOT be created when balance is 0
	txTxRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestAccountUsecase_Update_Success(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	userID := uuid.New()
	accountID := uuid.New()
	existingAccount := &domain.Account{
		ID:      accountID,
		UserID:  userID,
		Name:    "Old Name",
		Type:    domain.AccountTypeBank,
		Balance: 100000,
	}

	mockAccRepo.On("GetByID", mock.Anything, accountID, userID).Return(existingAccount, nil)
	mockAccRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)

	uc := usecase.NewAccountUsecase(mockAccRepo, mockTxRepo, nil)
	req := &domain.UpdateAccountRequest{Name: "New Name"}
	account, err := uc.Update(context.Background(), userID, accountID, req)

	require.NoError(t, err)
	assert.Equal(t, "New Name", account.Name)
	assert.Equal(t, int64(100000), account.Balance)
	mockAccRepo.AssertExpectations(t)
}

func TestAccountUsecase_Update_NotFound(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	userID := uuid.New()
	accountID := uuid.New()

	mockAccRepo.On("GetByID", mock.Anything, accountID, userID).Return(nil, errors.New("not found"))

	uc := usecase.NewAccountUsecase(mockAccRepo, mockTxRepo, nil)
	req := &domain.UpdateAccountRequest{Name: "New Name"}
	account, err := uc.Update(context.Background(), userID, accountID, req)

	assert.Nil(t, account)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	mockAccRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestAccountUsecase_SoftDelete_Success(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	userID := uuid.New()
	accountID := uuid.New()

	mockAccRepo.On("SoftDelete", mock.Anything, accountID, userID).Return(nil)

	uc := usecase.NewAccountUsecase(mockAccRepo, mockTxRepo, nil)
	err := uc.SoftDelete(context.Background(), userID, accountID)

	assert.NoError(t, err)
	mockAccRepo.AssertExpectations(t)
}

func TestAccountUsecase_FetchByUserID_Success(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	userID := uuid.New()
	expected := []domain.Account{
		{ID: uuid.New(), UserID: userID, Name: "Cash", Type: domain.AccountTypeCash, Balance: 100000},
		{ID: uuid.New(), UserID: userID, Name: "Bank", Type: domain.AccountTypeBank, Balance: 500000},
	}

	mockAccRepo.On("FetchByUserID", mock.Anything, userID).Return(expected, nil)

	uc := usecase.NewAccountUsecase(mockAccRepo, mockTxRepo, nil)
	accounts, err := uc.FetchByUserID(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, accounts, 2)
	assert.Equal(t, "Cash", accounts[0].Name)
	assert.Equal(t, "Bank", accounts[1].Name)
	mockAccRepo.AssertExpectations(t)
}
