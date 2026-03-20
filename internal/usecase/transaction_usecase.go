package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"gorm.io/gorm"
)

type transactionUsecase struct {
	txRepo  domain.TransactionRepository
	accRepo domain.AccountRepository
	catRepo domain.CategoryRepository
	db      *gorm.DB
}

func NewTransactionUsecase(txRepo domain.TransactionRepository, accRepo domain.AccountRepository, catRepo domain.CategoryRepository, db *gorm.DB) domain.TransactionUsecase {
	return &transactionUsecase{
		txRepo:  txRepo,
		accRepo: accRepo,
		catRepo: catRepo,
		db:      db,
	}
}

func (u *transactionUsecase) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateTransactionRequest) (*domain.CreateTransactionResponse, error) {
	slog.InfoContext(ctx, "Creating transaction", "user_id", userID, "account_id", req.AccountID, "type", req.TransactionType)

	if !isValidTxType(req.TransactionType) {
		return nil, domain.ErrInvalidTxType
	}

	var result *domain.CreateTransactionResponse

	err := u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := u.txRepo.WithTx(tx)
		accRepo := u.accRepo.WithTx(tx)

		account, err := accRepo.GetByID(ctx, req.AccountID, userID)
		if err != nil {
			return domain.ErrAccountNotFound
		}

		if req.CategoryID != nil {
			if _, err := u.catRepo.GetByID(ctx, *req.CategoryID, userID); err != nil {
				return domain.ErrCategoryNotFound
			}
		}

		transaction := &domain.Transaction{
			UserID:          userID,
			AccountID:       req.AccountID,
			CategoryID:      req.CategoryID,
			TransactionType: req.TransactionType,
			Amount:          req.Amount,
			Note:            req.Note,
			TransactionDate: req.TransactionDate,
		}
		if err := txRepo.Create(ctx, transaction); err != nil {
			slog.ErrorContext(ctx, "Failed to create transaction", "error", err)
			return err
		}

		delta := domain.CalculateBalanceDelta(req.TransactionType, req.Amount)
		if err := accRepo.UpdateBalance(ctx, req.AccountID, delta); err != nil {
			slog.ErrorContext(ctx, "Failed to update account balance", "error", err)
			return err
		}

		result = &domain.CreateTransactionResponse{
			Transaction:    *transaction,
			AccountBalance: account.Balance + delta,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Transaction created successfully", "transaction_id", result.Transaction.ID)
	return result, nil
}

func (u *transactionUsecase) Update(ctx context.Context, userID, txID uuid.UUID, req *domain.UpdateTransactionRequest) (*domain.UpdateTransactionResponse, error) {
	slog.InfoContext(ctx, "Updating transaction", "user_id", userID, "transaction_id", txID)

	var result *domain.UpdateTransactionResponse

	err := u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := u.txRepo.WithTx(tx)
		accRepo := u.accRepo.WithTx(tx)

		existing, err := txRepo.GetByID(ctx, txID, userID)
		if err != nil {
			return domain.ErrTransactionNotFound
		}

		if req.CategoryID != nil {
			if _, err := u.catRepo.GetByID(ctx, *req.CategoryID, userID); err != nil {
				return domain.ErrCategoryNotFound
			}
		}

		oldDelta := domain.CalculateBalanceDelta(existing.TransactionType, existing.Amount)
		newDelta := domain.CalculateBalanceDelta(existing.TransactionType, req.Amount)
		adjustmentDelta := newDelta - oldDelta

		existing.Amount = req.Amount
		existing.Note = req.Note
		existing.CategoryID = req.CategoryID
		existing.TransactionDate = req.TransactionDate

		if err := txRepo.Update(ctx, existing); err != nil {
			slog.ErrorContext(ctx, "Failed to update transaction", "error", err)
			return err
		}

		if adjustmentDelta != 0 {
			if err := accRepo.UpdateBalance(ctx, existing.AccountID, adjustmentDelta); err != nil {
				slog.ErrorContext(ctx, "Failed to update account balance", "error", err)
				return err
			}
		}

		account, err := accRepo.GetByID(ctx, existing.AccountID, userID)
		if err != nil {
			return err
		}

		result = &domain.UpdateTransactionResponse{
			Transaction:    *existing,
			AccountBalance: account.Balance,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Transaction updated successfully", "transaction_id", txID)
	return result, nil
}

func (u *transactionUsecase) SoftDelete(ctx context.Context, userID, txID uuid.UUID) error {
	slog.InfoContext(ctx, "Soft deleting transaction", "user_id", userID, "transaction_id", txID)

	err := u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := u.txRepo.WithTx(tx)
		accRepo := u.accRepo.WithTx(tx)

		existing, err := txRepo.GetByID(ctx, txID, userID)
		if err != nil {
			return domain.ErrTransactionNotFound
		}

		delta := domain.CalculateBalanceDelta(existing.TransactionType, existing.Amount)
		reversalDelta := -delta

		if err := txRepo.SoftDelete(ctx, txID, userID); err != nil {
			return err
		}

		if err := accRepo.UpdateBalance(ctx, existing.AccountID, reversalDelta); err != nil {
			slog.ErrorContext(ctx, "Failed to reverse account balance", "error", err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Transaction soft deleted successfully", "transaction_id", txID)
	return nil
}

func (u *transactionUsecase) Fetch(ctx context.Context, params domain.TransactionQueryParams) (*domain.PaginatedResult, error) {
	slog.InfoContext(ctx, "Fetching transactions", "user_id", params.UserID, "page", params.Page, "per_page", params.PerPage)

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	transactions, total, err := u.txRepo.Fetch(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch transactions", "error", err)
		return nil, err
	}

	totalPages := total / params.PerPage
	if total%params.PerPage != 0 {
		totalPages++
	}

	return &domain.PaginatedResult{
		Items:      transactions,
		TotalItems: total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

func isValidTxType(t domain.TransactionType) bool {
	switch t {
	case domain.TxTypeIncome, domain.TxTypeExpense, domain.TxTypeAdjustment:
		return true
	}
	return false
}
