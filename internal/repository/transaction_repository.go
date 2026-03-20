package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) WithTx(tx *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{db: tx}
}

func (r *transactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *transactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *transactionRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Transaction{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrTransactionNotFound
	}
	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&tx).Error
	if err != nil {
		return nil, domain.ErrTransactionNotFound
	}
	return &tx, nil
}

func (r *transactionRepository) Fetch(ctx context.Context, params domain.TransactionQueryParams) ([]domain.Transaction, int, error) {
	var transactions []domain.Transaction
	var total int64

	query := r.db.WithContext(ctx).
		Model(&domain.Transaction{}).
		Where("user_id = ? AND deleted_at IS NULL", params.UserID)

	if params.AccountID != nil {
		query = query.Where("account_id = ?", *params.AccountID)
	}
	if params.TransactionType != nil {
		query = query.Where("transaction_type = ?", *params.TransactionType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PerPage
	err := query.
		Order("transaction_date DESC").
		Limit(params.PerPage).
		Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, int(total), nil
}

func (r *transactionRepository) SumByAccount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var result int64
	err := r.db.WithContext(ctx).
		Model(&domain.Transaction{}).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Select(`COALESCE(SUM(CASE
			WHEN transaction_type IN ('INCOME', 'ADJUSTMENT') THEN amount
			WHEN transaction_type = 'EXPENSE' THEN -amount
			ELSE 0
		END), 0)`).
		Scan(&result).Error
	return result, err
}

func (r *transactionRepository) FetchRecent(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("transaction_date DESC").
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}
