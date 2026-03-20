package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"gorm.io/gorm"
)

type accountUsecase struct {
	accRepo domain.AccountRepository
	txRepo  domain.TransactionRepository
	db      *gorm.DB
}

func NewAccountUsecase(accRepo domain.AccountRepository, txRepo domain.TransactionRepository, db *gorm.DB) domain.AccountUsecase {
	return &accountUsecase{
		accRepo: accRepo,
		txRepo:  txRepo,
		db:      db,
	}
}

func (u *accountUsecase) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateAccountRequest) (*domain.Account, error) {
	slog.InfoContext(ctx, "Creating account", "user_id", userID, "name", req.Name, "type", req.Type)

	if !isValidAccountType(req.Type) {
		return nil, domain.ErrInvalidAccountType
	}
	if req.InitialBalance < 0 {
		return nil, domain.ErrNegativeBalance
	}

	var account domain.Account

	err := u.db.Transaction(func(tx *gorm.DB) error {
		accRepo := u.accRepo.WithTx(tx)
		txRepo := u.txRepo.WithTx(tx)

		account = domain.Account{
			UserID:  userID,
			Name:    req.Name,
			Type:    req.Type,
			Balance: req.InitialBalance,
		}
		if err := accRepo.Create(ctx, &account); err != nil {
			slog.ErrorContext(ctx, "Failed to create account", "error", err)
			return err
		}

		if req.InitialBalance > 0 {
			adjTx := &domain.Transaction{
				UserID:          userID,
				AccountID:       account.ID,
				TransactionType: domain.TxTypeAdjustment,
				Amount:          req.InitialBalance,
				Note:            "Initial balance",
				TransactionDate: time.Now(),
			}
			if err := txRepo.Create(ctx, adjTx); err != nil {
				slog.ErrorContext(ctx, "Failed to create adjustment transaction", "error", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Account created successfully", "account_id", account.ID)
	return &account, nil
}

func (u *accountUsecase) Update(ctx context.Context, userID, accountID uuid.UUID, req *domain.UpdateAccountRequest) (*domain.Account, error) {
	slog.InfoContext(ctx, "Updating account", "user_id", userID, "account_id", accountID)

	account, err := u.accRepo.GetByID(ctx, accountID, userID)
	if err != nil {
		return nil, domain.ErrAccountNotFound
	}

	account.Name = req.Name
	if err := u.accRepo.Update(ctx, account); err != nil {
		slog.ErrorContext(ctx, "Failed to update account", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Account updated successfully", "account_id", accountID)
	return account, nil
}

func (u *accountUsecase) SoftDelete(ctx context.Context, userID, accountID uuid.UUID) error {
	slog.InfoContext(ctx, "Soft deleting account", "user_id", userID, "account_id", accountID)

	if err := u.accRepo.SoftDelete(ctx, accountID, userID); err != nil {
		return err
	}

	slog.InfoContext(ctx, "Account soft deleted successfully", "account_id", accountID)
	return nil
}

func (u *accountUsecase) FetchByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	slog.InfoContext(ctx, "Fetching accounts", "user_id", userID)
	return u.accRepo.FetchByUserID(ctx, userID)
}

func isValidAccountType(t domain.AccountType) bool {
	switch t {
	case domain.AccountTypeCash, domain.AccountTypeBank, domain.AccountTypeEWallet, domain.AccountTypeCreditCard:
		return true
	}
	return false
}
