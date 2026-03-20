package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCategoryUsecase_BulkCreate_Success(t *testing.T) {
	db, smock := newTestDB(t)
	mockCatRepo := new(mocks.MockCategoryRepository)

	txCatRepo := new(mocks.MockCategoryRepository)
	mockCatRepo.On("WithTx", mock.Anything).Return(txCatRepo)

	userID := uuid.New()
	req := &domain.BulkCreateCategoryRequest{
		Categories: []domain.BulkCreateCategoryItem{
			{Name: "Gaji", Type: domain.CategoryTypeIncome, Icon: "💰"},
			{Name: "Makan", Type: domain.CategoryTypeExpense, Icon: "🍔"},
			{Name: "Transport", Type: domain.CategoryTypeExpense, Icon: "🚗"},
		},
	}

	txCatRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Category")).Return(nil)

	smock.ExpectBegin()
	smock.ExpectCommit()

	uc := usecase.NewCategoryUsecase(mockCatRepo, db)
	categories, err := uc.BulkCreate(context.Background(), userID, req)

	require.NoError(t, err)
	require.Len(t, categories, 3)
	assert.Equal(t, "Gaji", categories[0].Name)
	assert.Equal(t, domain.CategoryTypeIncome, categories[0].Type)
	assert.Equal(t, "Makan", categories[1].Name)
	assert.Equal(t, domain.CategoryTypeExpense, categories[1].Type)
	assert.Equal(t, "Transport", categories[2].Name)

	txCatRepo.AssertNumberOfCalls(t, "Create", 3)
	require.NoError(t, smock.ExpectationsWereMet())
}

func TestCategoryUsecase_BulkCreate_EmptyList(t *testing.T) {
	db, _ := newTestDB(t)
	mockCatRepo := new(mocks.MockCategoryRepository)

	userID := uuid.New()
	req := &domain.BulkCreateCategoryRequest{
		Categories: []domain.BulkCreateCategoryItem{},
	}

	uc := usecase.NewCategoryUsecase(mockCatRepo, db)
	categories, err := uc.BulkCreate(context.Background(), userID, req)

	assert.Nil(t, categories)
	assert.ErrorIs(t, err, domain.ErrEmptyBulkRequest)

	mockCatRepo.AssertNotCalled(t, "WithTx", mock.Anything)
}

func TestCategoryUsecase_BulkCreate_InvalidItem(t *testing.T) {
	db, _ := newTestDB(t)
	mockCatRepo := new(mocks.MockCategoryRepository)

	userID := uuid.New()
	req := &domain.BulkCreateCategoryRequest{
		Categories: []domain.BulkCreateCategoryItem{
			{Name: "Gaji", Type: domain.CategoryTypeIncome},
			{Name: "", Type: domain.CategoryTypeExpense}, // invalid: empty name
		},
	}

	uc := usecase.NewCategoryUsecase(mockCatRepo, db)
	categories, err := uc.BulkCreate(context.Background(), userID, req)

	assert.Nil(t, categories)
	assert.ErrorIs(t, err, domain.ErrValidationFailed)
	assert.Contains(t, err.Error(), "category[1]")

	mockCatRepo.AssertNotCalled(t, "WithTx", mock.Anything)
}

func TestCategoryUsecase_FetchByUserID_WithFilter(t *testing.T) {
	db, _ := newTestDB(t)
	mockCatRepo := new(mocks.MockCategoryRepository)

	userID := uuid.New()
	filterType := domain.CategoryTypeExpense

	expected := []domain.Category{
		{ID: uuid.New(), UserID: userID, Name: "Makan", Type: domain.CategoryTypeExpense},
		{ID: uuid.New(), UserID: userID, Name: "Transport", Type: domain.CategoryTypeExpense},
	}

	mockCatRepo.On("FetchByUserID", mock.Anything, userID, &filterType).Return(expected, nil)

	uc := usecase.NewCategoryUsecase(mockCatRepo, db)
	categories, err := uc.FetchByUserID(context.Background(), userID, &filterType)

	require.NoError(t, err)
	require.Len(t, categories, 2)
	assert.Equal(t, "Makan", categories[0].Name)
	assert.Equal(t, domain.CategoryTypeExpense, categories[0].Type)
	assert.Equal(t, "Transport", categories[1].Name)

	mockCatRepo.AssertExpectations(t)
}
