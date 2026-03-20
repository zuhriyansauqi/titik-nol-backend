package usecase

import (
	"context"
	"log/slog"

	"github.com/mzhryns/titik-nol-backend/internal/domain"
)

type ReconciliationService struct {
	accRepo domain.AccountRepository
	txRepo  domain.TransactionRepository
}

func NewReconciliationService(accRepo domain.AccountRepository, txRepo domain.TransactionRepository) *ReconciliationService {
	return &ReconciliationService{
		accRepo: accRepo,
		txRepo:  txRepo,
	}
}

func (s *ReconciliationService) ReconcileAll(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting balance reconciliation for all accounts")

	accounts, err := s.accRepo.GetAllActive(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch active accounts for reconciliation", "error", err)
		return err
	}

	for _, account := range accounts {
		s.ReconcileAccount(ctx, account)
	}

	slog.InfoContext(ctx, "Balance reconciliation completed", "accounts_checked", len(accounts))
	return nil
}

func (s *ReconciliationService) ReconcileAccount(ctx context.Context, account domain.Account) {
	expectedBalance, err := s.txRepo.SumByAccount(ctx, account.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to calculate expected balance", "account_id", account.ID, "error", err)
		return
	}

	if expectedBalance != account.Balance {
		slog.WarnContext(ctx, "Balance mismatch detected",
			"account_id", account.ID,
			"expected_balance", expectedBalance,
			"stored_balance", account.Balance,
		)
	}
}
