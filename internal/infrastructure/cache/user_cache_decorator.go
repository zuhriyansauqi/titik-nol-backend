package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

// UserCacheDecorator is a cache decorator for domain.UserRepository.
type UserCacheDecorator struct {
	repo         domain.UserRepository
	redis        *RedisClient
	ttl          time.Duration
	cacheEnabled bool
}

// NewUserCacheDecorator wraps a UserRepository with Redis caching.
func NewUserCacheDecorator(
	repo domain.UserRepository,
	redisClient *RedisClient,
	cfg *config.Config,
) domain.UserRepository {
	return &UserCacheDecorator{
		repo:         repo,
		redis:        redisClient,
		ttl:          cfg.CacheUserTTL,
		cacheEnabled: cfg.CacheEnabled,
	}
}

// GetByID checks cache first, falls back to PG on miss.
func (c *UserCacheDecorator) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return c.repo.GetByID(ctx, id)
	}

	key := c.redis.BuildKey("user", id.String())

	var user domain.User
	err := c.redis.Get(ctx, key, &user)
	if err == nil {
		return &user, nil
	}
	if err != redis.Nil {
		slog.ErrorContext(ctx, "cache get failed for user", "key", key, "error", err)
	}

	result, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if setErr := c.redis.Set(ctx, key, result, c.ttl); setErr != nil {
		slog.ErrorContext(ctx, "cache set failed for user", "key", key, "error", setErr)
	}

	return result, nil
}

// Update delegates to PG, then invalidates the user cache key.
func (c *UserCacheDecorator) Update(ctx context.Context, user *domain.User) error {
	if err := c.repo.Update(ctx, user); err != nil {
		return err
	}

	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return nil
	}

	key := c.redis.BuildKey("user", user.ID.String())
	if err := c.redis.Delete(ctx, key); err != nil {
		slog.ErrorContext(ctx, "cache invalidation failed for user", "user_id", user.ID, "error", err)
	}

	return nil
}

// Create delegates directly to PG (no caching).
func (c *UserCacheDecorator) Create(ctx context.Context, user *domain.User) error {
	return c.repo.Create(ctx, user)
}

// GetByEmail delegates directly to PG (no caching).
func (c *UserCacheDecorator) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return c.repo.GetByEmail(ctx, email)
}

// GetByProviderID delegates directly to PG (no caching).
func (c *UserCacheDecorator) GetByProviderID(ctx context.Context, providerID string) (*domain.User, error) {
	return c.repo.GetByProviderID(ctx, providerID)
}

// Fetch delegates directly to PG (no caching).
func (c *UserCacheDecorator) Fetch(ctx context.Context, params domain.PaginationParams) ([]domain.User, int, error) {
	return c.repo.Fetch(ctx, params)
}
