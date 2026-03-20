package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockCategoryUsecase struct {
	mock.Mock
}

func (m *MockCategoryUsecase) BulkCreate(ctx context.Context, userID uuid.UUID, req *domain.BulkCreateCategoryRequest) ([]domain.Category, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *MockCategoryUsecase) FetchByUserID(ctx context.Context, userID uuid.UUID, filterType *domain.CategoryType) ([]domain.Category, error) {
	args := m.Called(ctx, userID, filterType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}
