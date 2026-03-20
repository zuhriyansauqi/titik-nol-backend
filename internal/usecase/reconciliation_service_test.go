package usecase_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReconciliationService_ReconcileAccount_Match(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	accountID := uuid.New()
	account := domain.Account{
		ID:      accountID,
		UserID:  uuid.New(),
		Name:    "Cash",
		Type:    domain.AccountTypeCash,
		Balance: 500000,
	}

	mockTxRepo.On("SumByAccount", mock.Anything, accountID).Return(int64(500000), nil)

	// Capture log output to verify no warning is emitted
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	original := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(original)

	svc := usecase.NewReconciliationService(mockAccRepo, mockTxRepo)
	svc.ReconcileAccount(context.Background(), account)

	assert.NotContains(t, buf.String(), "Balance mismatch detected")

	mockTxRepo.AssertExpectations(t)
}

func TestReconciliationService_ReconcileAccount_Mismatch(t *testing.T) {
	mockAccRepo := new(mocks.MockAccountRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)

	accountID := uuid.New()
	account := domain.Account{
		ID:      accountID,
		UserID:  uuid.New(),
		Name:    "Bank BCA",
		Type:    domain.AccountTypeBank,
		Balance: 300000, // stored balance
	}

	// Expected balance from transactions differs from stored balance
	mockTxRepo.On("SumByAccount", mock.Anything, accountID).Return(int64(250000), nil)

	// Capture log output to verify warning is emitted
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	original := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(original)

	svc := usecase.NewReconciliationService(mockAccRepo, mockTxRepo)
	svc.ReconcileAccount(context.Background(), account)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Balance mismatch detected")
	assert.Contains(t, logOutput, accountID.String())

	mockTxRepo.AssertExpectations(t)
}
