package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockDashboardUsecase struct {
	mock.Mock
}

func (m *MockDashboardUsecase) GetSummary(ctx context.Context, userID uuid.UUID) (*domain.DashboardSummary, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DashboardSummary), args.Error(1)
}
