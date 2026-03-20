package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockAccountUsecase struct {
	mock.Mock
}

func (m *MockAccountUsecase) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateAccountRequest) (*domain.Account, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountUsecase) Update(ctx context.Context, userID, accountID uuid.UUID, req *domain.UpdateAccountRequest) (*domain.Account, error) {
	args := m.Called(ctx, userID, accountID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountUsecase) SoftDelete(ctx context.Context, userID, accountID uuid.UUID) error {
	args := m.Called(ctx, userID, accountID)
	return args.Error(0)
}

func (m *MockAccountUsecase) FetchByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Account), args.Error(1)
}
