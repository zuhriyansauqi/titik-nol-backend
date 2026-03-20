package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "INCOME"
	CategoryTypeExpense CategoryType = "EXPENSE"
)

type Category struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	Name      string       `gorm:"size:100;not null" json:"name"`
	Type      CategoryType `gorm:"type:category_type_enum;not null" json:"type"`
	Icon      string       `gorm:"size:50" json:"icon,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// CategoryRepository defines the data access interface for categories.
type CategoryRepository interface {
	WithTx(tx *gorm.DB) CategoryRepository
	Create(ctx context.Context, category *Category) error
	FetchByUserID(ctx context.Context, userID uuid.UUID, filterType *CategoryType) ([]Category, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Category, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// CategoryUsecase defines the business logic interface for categories.
type CategoryUsecase interface {
	BulkCreate(ctx context.Context, userID uuid.UUID, req *BulkCreateCategoryRequest) ([]Category, error)
	FetchByUserID(ctx context.Context, userID uuid.UUID, filterType *CategoryType) ([]Category, error)
}

// BulkCreateCategoryItem is a single item in a bulk category creation request.
type BulkCreateCategoryItem struct {
	Name string       `json:"name" binding:"required"`
	Type CategoryType `json:"type" binding:"required"`
	Icon string       `json:"icon"`
}

// BulkCreateCategoryRequest is the DTO for bulk creating categories.
type BulkCreateCategoryRequest struct {
	Categories []BulkCreateCategoryItem `json:"categories" binding:"required,min=1,dive"`
}
