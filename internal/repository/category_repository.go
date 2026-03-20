package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) WithTx(tx *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: tx}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) FetchByUserID(ctx context.Context, userID uuid.UUID, filterType *domain.CategoryType) ([]domain.Category, error) {
	var categories []domain.Category
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if filterType != nil {
		query = query.Where("type = ?", *filterType)
	}
	err := query.Order("created_at DESC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&category).Error
	if err != nil {
		return nil, domain.ErrCategoryNotFound
	}
	return &category, nil
}

func (r *categoryRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Category{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
