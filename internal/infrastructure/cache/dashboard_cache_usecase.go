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

// DashboardCacheUsecase is a cache decorator for domain.DashboardUsecase.
type DashboardCacheUsecase struct {
	usecase      domain.DashboardUsecase
	redis        *RedisClient
	ttl          time.Duration
	cacheEnabled bool
}

// NewDashboardCacheUsecase wraps a DashboardUsecase with Redis caching.
func NewDashboardCacheUsecase(
	usecase domain.DashboardUsecase,
	redisClient *RedisClient,
	cfg *config.Config,
) domain.DashboardUsecase {
	return &DashboardCacheUsecase{
		usecase:      usecase,
		redis:        redisClient,
		ttl:          cfg.CacheDashboardTTL,
		cacheEnabled: cfg.CacheEnabled,
	}
}

// GetSummary checks cache first, falls back to underlying usecase on miss.
func (c *DashboardCacheUsecase) GetSummary(ctx context.Context, userID uuid.UUID) (*domain.DashboardSummary, error) {
	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return c.usecase.GetSummary(ctx, userID)
	}

	key := c.redis.BuildKey("dashboard", userID.String())

	var summary domain.DashboardSummary
	err := c.redis.Get(ctx, key, &summary)
	if err == nil {
		return &summary, nil
	}
	if err != redis.Nil {
		slog.ErrorContext(ctx, "cache get failed for dashboard", "key", key, "error", err)
	}

	result, err := c.usecase.GetSummary(ctx, userID)
	if err != nil {
		return nil, err
	}

	if setErr := c.redis.Set(ctx, key, result, c.ttl); setErr != nil {
		slog.ErrorContext(ctx, "cache set failed for dashboard", "key", key, "error", setErr)
	}

	return result, nil
}
