package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/api/idtoken"
)

type MockGoogleSSOService struct {
	mock.Mock
}

func (m *MockGoogleSSOService) VerifyIDToken(ctx context.Context, idToken string) (*idtoken.Payload, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idtoken.Payload), args.Error(1)
}
