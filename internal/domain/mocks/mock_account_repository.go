package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) WithTx(tx *gorm.DB) domain.AccountRepository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(domain.AccountRepository)
}

func (m *MockAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepository) Update(ctx context.Context, account *domain.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAccountRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Account, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) FetchByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Account), args.Error(1)
}

func (m *MockAccountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta int64) error {
	args := m.Called(ctx, id, delta)
	return args.Error(0)
}

func (m *MockAccountRepository) GetAllActive(ctx context.Context) ([]domain.Account, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Account), args.Error(1)
}
