package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	user := &domain.User{Email: "new@example.com", Name: "New User"}

	mockRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
	mockRepo.On("Create", mock.Anything, user).Return(nil)

	err := uc.Create(context.Background(), user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreate_DuplicateEmail(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	existingUser := &domain.User{Email: "exists@example.com", Name: "Existing"}
	newUser := &domain.User{Email: "exists@example.com", Name: "New"}

	mockRepo.On("GetByEmail", mock.Anything, "exists@example.com").Return(existingUser, nil)

	err := uc.Create(context.Background(), newUser)

	assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestGetByID_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	id := uuid.New()
	expected := &domain.User{ID: id, Email: "test@example.com", Name: "Test"}

	mockRepo.On("GetByID", mock.Anything, id).Return(expected, nil)

	user, err := uc.GetByID(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, expected.Email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	id := uuid.New()
	mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	user, err := uc.GetByID(context.Background(), id)

	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestFetch_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	params := domain.PaginationParams{Page: 1, PerPage: 10}
	users := []domain.User{
		{ID: uuid.New(), Email: "a@example.com"},
		{ID: uuid.New(), Email: "b@example.com"},
	}

	mockRepo.On("Fetch", mock.Anything, params).Return(users, 25, nil)

	result, err := uc.Fetch(context.Background(), params)

	require.NoError(t, err)
	assert.Equal(t, 25, result.TotalItems)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.Equal(t, 3, result.TotalPages) // 25/10 = 2.5 → 3
	mockRepo.AssertExpectations(t)
}
