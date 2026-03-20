package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"gorm.io/gorm"
)

type onboardingUsecase struct {
	accRepo domain.AccountRepository
	txRepo  domain.TransactionRepository
	db      *gorm.DB
}

func NewOnboardingUsecase(accRepo domain.AccountRepository, txRepo domain.TransactionRepository, db *gorm.DB) domain.OnboardingUsecase {
	return &onboardingUsecase{
		accRepo: accRepo,
		txRepo:  txRepo,
		db:      db,
	}
}

func (u *onboardingUsecase) SetupAccounts(ctx context.Context, userID uuid.UUID, req *domain.SetupAccountsRequest) (*domain.SetupAccountsResponse, error) {
	slog.InfoContext(ctx, "Setting up onboarding accounts", "user_id", userID, "count", len(req.Accounts))

	if len(req.Accounts) == 0 {
		return nil, domain.ErrEmptyBulkRequest
	}

	for i, item := range req.Accounts {
		if item.Name == "" {
			return nil, fmt.Errorf("account[%d]: name is required: %w", i, domain.ErrValidationFailed)
		}
		if !isValidAccountType(item.Type) {
			return nil, fmt.Errorf("account[%d]: %w", i, domain.ErrInvalidAccountType)
		}
		if item.InitialBalance < 0 {
			return nil, fmt.Errorf("account[%d]: %w", i, domain.ErrNegativeBalance)
		}
	}

	var resp domain.SetupAccountsResponse

	err := u.db.Transaction(func(tx *gorm.DB) error {
		accRepo := u.accRepo.WithTx(tx)
		txRepo := u.txRepo.WithTx(tx)

		for _, item := range req.Accounts {
			account := &domain.Account{
				UserID:  userID,
				Name:    item.Name,
				Type:    item.Type,
				Balance: item.InitialBalance,
			}
			if err := accRepo.Create(ctx, account); err != nil {
				slog.ErrorContext(ctx, "Failed to create account during onboarding", "error", err)
				return err
			}
			resp.Accounts = append(resp.Accounts, *account)

			if item.InitialBalance > 0 {
				adjTx := &domain.Transaction{
					UserID:          userID,
					AccountID:       account.ID,
					TransactionType: domain.TxTypeAdjustment,
					Amount:          item.InitialBalance,
					Note:            "Saldo awal (onboarding)",
					TransactionDate: time.Now(),
				}
				if err := txRepo.Create(ctx, adjTx); err != nil {
					slog.ErrorContext(ctx, "Failed to create adjustment transaction during onboarding", "error", err)
					return err
				}
				resp.Transactions = append(resp.Transactions, *adjTx)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Onboarding accounts setup completed", "user_id", userID, "accounts_created", len(resp.Accounts))
	return &resp, nil
}
