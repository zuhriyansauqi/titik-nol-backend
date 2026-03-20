package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"gorm.io/gorm"
)

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) domain.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) WithTx(tx *gorm.DB) domain.AccountRepository {
	return &accountRepository{db: tx}
}

func (r *accountRepository) Create(ctx context.Context, account *domain.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *accountRepository) Update(ctx context.Context, account *domain.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *accountRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Account{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *accountRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Account, error) {
	var account domain.Account
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&account).Error
	if err != nil {
		return nil, domain.ErrAccountNotFound
	}
	return &account, nil
}

func (r *accountRepository) FetchByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	var accounts []domain.Account
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta int64) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Account{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("balance", gorm.Expr("balance + ?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *accountRepository) GetAllActive(ctx context.Context) ([]domain.Account, error) {
	var accounts []domain.Account
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Find(&accounts).Error
	return accounts, err
}
