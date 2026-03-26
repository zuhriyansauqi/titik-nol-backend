package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID uuid.UUID, role string) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (uuid.UUID, string, error) {
	args := m.Called(tokenString)
	return args.Get(0).(uuid.UUID), args.String(1), args.Error(2)
}
