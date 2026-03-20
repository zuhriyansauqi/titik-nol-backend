package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountType string

const (
	AccountTypeCash       AccountType = "CASH"
	AccountTypeBank       AccountType = "BANK"
	AccountTypeEWallet    AccountType = "E_WALLET"
	AccountTypeCreditCard AccountType = "CREDIT_CARD"
)

type Account struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null" json:"user_id"`
	Name      string      `gorm:"size:100;not null" json:"name"`
	Type      AccountType `gorm:"type:account_type_enum;not null" json:"type"`
	Balance   int64       `gorm:"default:0" json:"balance"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	DeletedAt *time.Time  `gorm:"index" json:"deleted_at,omitempty"`
}

type AccountRepository interface {
	WithTx(tx *gorm.DB) AccountRepository
	Create(ctx context.Context, account *Account) error
	Update(ctx context.Context, account *Account) error
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Account, error)
	FetchByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error)
	UpdateBalance(ctx context.Context, id uuid.UUID, delta int64) error
	GetAllActive(ctx context.Context) ([]Account, error)
}

type AccountUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, req *CreateAccountRequest) (*Account, error)
	Update(ctx context.Context, userID, accountID uuid.UUID, req *UpdateAccountRequest) (*Account, error)
	SoftDelete(ctx context.Context, userID, accountID uuid.UUID) error
	FetchByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error)
}

type CreateAccountRequest struct {
	Name           string      `json:"name" binding:"required"`
	Type           AccountType `json:"type" binding:"required"`
	InitialBalance int64       `json:"initial_balance" binding:"min=0"`
}

type UpdateAccountRequest struct {
	Name string `json:"name" binding:"required"`
}
