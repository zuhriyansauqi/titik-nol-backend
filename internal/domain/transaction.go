package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TxTypeIncome     TransactionType = "INCOME"
	TxTypeExpense    TransactionType = "EXPENSE"
	TxTypeTransfer   TransactionType = "TRANSFER"
	TxTypeAdjustment TransactionType = "ADJUSTMENT"
)

type Transaction struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID          uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	AccountID       uuid.UUID       `gorm:"type:uuid;not null" json:"account_id"`
	CategoryID      *uuid.UUID      `gorm:"type:uuid" json:"category_id,omitempty"`
	TransactionType TransactionType `gorm:"type:tx_type_enum;not null" json:"transaction_type"`
	Amount          int64           `gorm:"not null" json:"amount"`
	Note            string          `json:"note,omitempty"`
	TransactionDate time.Time       `gorm:"not null" json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
	DeletedAt       *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
}

type TransactionQueryParams struct {
	UserID          uuid.UUID
	AccountID       *uuid.UUID
	TransactionType *TransactionType
	Page            int
	PerPage         int
}

// TransactionRepository defines the data access interface for transactions.
type TransactionRepository interface {
	WithTx(tx *gorm.DB) TransactionRepository
	Create(ctx context.Context, tx *Transaction) error
	Update(ctx context.Context, tx *Transaction) error
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Transaction, error)
	Fetch(ctx context.Context, params TransactionQueryParams) ([]Transaction, int, error)
	SumByAccount(ctx context.Context, accountID uuid.UUID) (int64, error)
	FetchRecent(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error)
}

// TransactionUsecase defines the business logic interface for transactions.
type TransactionUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, req *CreateTransactionRequest) (*CreateTransactionResponse, error)
	Update(ctx context.Context, userID, txID uuid.UUID, req *UpdateTransactionRequest) (*UpdateTransactionResponse, error)
	SoftDelete(ctx context.Context, userID, txID uuid.UUID) error
	Fetch(ctx context.Context, params TransactionQueryParams) (*PaginatedResult, error)
}

// CreateTransactionRequest is the DTO for creating a new transaction.
type CreateTransactionRequest struct {
	AccountID       uuid.UUID       `json:"account_id" binding:"required"`
	CategoryID      *uuid.UUID      `json:"category_id"`
	TransactionType TransactionType `json:"transaction_type" binding:"required"`
	Amount          int64           `json:"amount" binding:"required,gt=0"`
	Note            string          `json:"note"`
	TransactionDate time.Time       `json:"transaction_date" binding:"required"`
}

// CreateTransactionResponse is the DTO returned after creating a transaction.
type CreateTransactionResponse struct {
	Transaction    Transaction `json:"transaction"`
	AccountBalance int64       `json:"account_balance"`
}

// UpdateTransactionRequest is the DTO for updating an existing transaction.
type UpdateTransactionRequest struct {
	Amount          int64      `json:"amount" binding:"required,gt=0"`
	Note            string     `json:"note"`
	CategoryID      *uuid.UUID `json:"category_id"`
	TransactionDate time.Time  `json:"transaction_date" binding:"required"`
}

// UpdateTransactionResponse is the DTO returned after updating a transaction.
type UpdateTransactionResponse struct {
	Transaction    Transaction `json:"transaction"`
	AccountBalance int64       `json:"account_balance"`
}

// CalculateBalanceDelta returns the balance delta for a given transaction type and amount.
// INCOME and ADJUSTMENT add to balance, EXPENSE subtracts, unknown types return 0.
func CalculateBalanceDelta(txType TransactionType, amount int64) int64 {
	switch txType {
	case TxTypeIncome, TxTypeAdjustment:
		return amount
	case TxTypeExpense:
		return -amount
	default:
		return 0
	}
}
