package cache

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CategoryCacheDecorator is a cache decorator for domain.CategoryRepository.
type CategoryCacheDecorator struct {
	repo         domain.CategoryRepository
	redis        *RedisClient
	ttl          time.Duration
	cacheEnabled bool
}

// NewCategoryCacheDecorator wraps a CategoryRepository with Redis caching.
func NewCategoryCacheDecorator(
	repo domain.CategoryRepository,
	redisClient *RedisClient,
	cfg *config.Config,
) domain.CategoryRepository {
	return &CategoryCacheDecorator{
		repo:         repo,
		redis:        redisClient,
		ttl:          cfg.CacheCategoryTTL,
		cacheEnabled: cfg.CacheEnabled,
	}
}

// WithTx returns the underlying repo's WithTx — cache is bypassed inside DB transactions.
func (c *CategoryCacheDecorator) WithTx(tx *gorm.DB) domain.CategoryRepository {
	return c.repo.WithTx(tx)
}

// FetchByUserID checks cache first, falls back to PG on miss.
func (c *CategoryCacheDecorator) FetchByUserID(ctx context.Context, userID uuid.UUID, filterType *domain.CategoryType) ([]domain.Category, error) {
	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return c.repo.FetchByUserID(ctx, userID, filterType)
	}

	filter := "all"
	if filterType != nil {
		filter = strings.ToLower(string(*filterType))
	}
	key := c.redis.BuildKey("category", "list", userID.String(), filter)

	var categories []domain.Category
	err := c.redis.Get(ctx, key, &categories)
	if err == nil {
		return categories, nil
	}
	if err != redis.Nil {
		slog.ErrorContext(ctx, "cache get failed for category list", "key", key, "error", err)
	}

	categories, err = c.repo.FetchByUserID(ctx, userID, filterType)
	if err != nil {
		return nil, err
	}

	if setErr := c.redis.Set(ctx, key, categories, c.ttl); setErr != nil {
		slog.ErrorContext(ctx, "cache set failed for category list", "key", key, "error", setErr)
	}

	return categories, nil
}

// GetByID checks cache first, falls back to PG on miss.
func (c *CategoryCacheDecorator) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Category, error) {
	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return c.repo.GetByID(ctx, id, userID)
	}

	key := c.redis.BuildKey("category", userID.String(), id.String())

	var category domain.Category
	err := c.redis.Get(ctx, key, &category)
	if err == nil {
		return &category, nil
	}
	if err != redis.Nil {
		slog.ErrorContext(ctx, "cache get failed for category", "key", key, "error", err)
	}

	result, err := c.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if setErr := c.redis.Set(ctx, key, result, c.ttl); setErr != nil {
		slog.ErrorContext(ctx, "cache set failed for category", "key", key, "error", setErr)
	}

	return result, nil
}

// Create delegates to PG, then invalidates all category cache entries for the user.
func (c *CategoryCacheDecorator) Create(ctx context.Context, category *domain.Category) error {
	if err := c.repo.Create(ctx, category); err != nil {
		return err
	}

	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return nil
	}

	pattern := c.redis.BuildKey("category", "*"+category.UserID.String()+"*")
	if err := c.redis.DeleteByPattern(ctx, pattern); err != nil {
		slog.ErrorContext(ctx, "cache invalidation failed for categories", "user_id", category.UserID, "error", err)
	}

	return nil
}

// CountByUserID always delegates to PG (not cached).
func (c *CategoryCacheDecorator) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return c.repo.CountByUserID(ctx, userID)
}
