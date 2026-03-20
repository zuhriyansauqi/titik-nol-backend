package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"gorm.io/gorm"
)

type categoryUsecase struct {
	catRepo domain.CategoryRepository
	db      *gorm.DB
}

func NewCategoryUsecase(catRepo domain.CategoryRepository, db *gorm.DB) domain.CategoryUsecase {
	return &categoryUsecase{
		catRepo: catRepo,
		db:      db,
	}
}

func (u *categoryUsecase) BulkCreate(ctx context.Context, userID uuid.UUID, req *domain.BulkCreateCategoryRequest) ([]domain.Category, error) {
	slog.InfoContext(ctx, "Bulk creating categories", "user_id", userID, "count", len(req.Categories))

	if len(req.Categories) == 0 {
		return nil, domain.ErrEmptyBulkRequest
	}

	for i, item := range req.Categories {
		if item.Name == "" {
			return nil, fmt.Errorf("category[%d]: name is required: %w", i, domain.ErrValidationFailed)
		}
		if !isValidCategoryType(item.Type) {
			return nil, fmt.Errorf("category[%d]: %w", i, domain.ErrInvalidCategoryType)
		}
	}

	var categories []domain.Category

	err := u.db.Transaction(func(tx *gorm.DB) error {
		catRepo := u.catRepo.WithTx(tx)

		for _, item := range req.Categories {
			cat := &domain.Category{
				UserID: userID,
				Name:   item.Name,
				Type:   item.Type,
				Icon:   item.Icon,
			}
			if err := catRepo.Create(ctx, cat); err != nil {
				slog.ErrorContext(ctx, "Failed to create category", "error", err)
				return err
			}
			categories = append(categories, *cat)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Categories created successfully", "user_id", userID, "count", len(categories))
	return categories, nil
}

func (u *categoryUsecase) FetchByUserID(ctx context.Context, userID uuid.UUID, filterType *domain.CategoryType) ([]domain.Category, error) {
	slog.InfoContext(ctx, "Fetching categories", "user_id", userID)
	return u.catRepo.FetchByUserID(ctx, userID, filterType)
}

func isValidCategoryType(t domain.CategoryType) bool {
	switch t {
	case domain.CategoryTypeIncome, domain.CategoryTypeExpense:
		return true
	}
	return false
}
