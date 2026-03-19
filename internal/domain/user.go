package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuthProvider string

const (
	ProviderGoogle AuthProvider = "GOOGLE"
)

type User struct {
	ID         uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email      string       `gorm:"size:255;not null;unique" json:"email"`
	Name       string       `gorm:"size:255;not null" json:"name"`
	AvatarURL  string       `gorm:"column:avatar_url" json:"avatar_url"`
	Provider   AuthProvider `gorm:"size:50;not null" json:"provider"`
	ProviderID string       `gorm:"column:provider_id;size:255;uniqueIndex;not null" json:"provider_id"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// PaginatedResult wraps a paginated response with metadata.
type PaginatedResult struct {
	Items      any `json:"items"`
	TotalItems int `json:"total_items"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByProviderID(ctx context.Context, providerID string) (*User, error)
	Fetch(ctx context.Context, params PaginationParams) ([]User, int, error)
}

type UserUsecase interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Fetch(ctx context.Context, params PaginationParams) (*PaginatedResult, error)
}

