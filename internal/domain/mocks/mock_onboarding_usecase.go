package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockOnboardingUsecase struct {
	mock.Mock
}

func (m *MockOnboardingUsecase) SetupAccounts(ctx context.Context, userID uuid.UUID, req *domain.SetupAccountsRequest) (*domain.SetupAccountsResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SetupAccountsResponse), args.Error(1)
}
