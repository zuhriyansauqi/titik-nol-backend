package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockTransactionUsecase struct {
	mock.Mock
}

func (m *MockTransactionUsecase) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateTransactionRequest) (*domain.CreateTransactionResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CreateTransactionResponse), args.Error(1)
}

func (m *MockTransactionUsecase) Update(ctx context.Context, userID, txID uuid.UUID, req *domain.UpdateTransactionRequest) (*domain.UpdateTransactionResponse, error) {
	args := m.Called(ctx, userID, txID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UpdateTransactionResponse), args.Error(1)
}

func (m *MockTransactionUsecase) SoftDelete(ctx context.Context, userID, txID uuid.UUID) error {
	args := m.Called(ctx, userID, txID)
	return args.Error(0)
}

func (m *MockTransactionUsecase) Fetch(ctx context.Context, params domain.TransactionQueryParams) (*domain.PaginatedResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PaginatedResult), args.Error(1)
}
