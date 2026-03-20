package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
)

type dashboardUsecase struct {
	accRepo domain.AccountRepository
	txRepo  domain.TransactionRepository
	catRepo domain.CategoryRepository
}

func NewDashboardUsecase(accRepo domain.AccountRepository, txRepo domain.TransactionRepository, catRepo domain.CategoryRepository) domain.DashboardUsecase {
	return &dashboardUsecase{
		accRepo: accRepo,
		txRepo:  txRepo,
		catRepo: catRepo,
	}
}

func (u *dashboardUsecase) GetSummary(ctx context.Context, userID uuid.UUID) (*domain.DashboardSummary, error) {
	slog.InfoContext(ctx, "Fetching dashboard summary", "user_id", userID)

	accounts, err := u.accRepo.FetchByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch accounts for dashboard", "error", err)
		return nil, err
	}

	var totalBalance int64
	for _, acc := range accounts {
		totalBalance += acc.Balance
	}

	recentTx, err := u.txRepo.FetchRecent(ctx, userID, 5)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch recent transactions for dashboard", "error", err)
		return nil, err
	}

	catCount, err := u.catRepo.CountByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to count categories for dashboard", "error", err)
		return nil, err
	}

	return &domain.DashboardSummary{
		TotalBalance:       totalBalance,
		RecentTransactions: recentTx,
		NeedsPaydaySetup:   catCount == 0,
	}, nil
}
