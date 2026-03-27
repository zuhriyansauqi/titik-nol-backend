package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AccountCacheDecorator is a cache decorator for domain.AccountRepository.
type AccountCacheDecorator struct {
	repo         domain.AccountRepository
	redis        *RedisClient
	ttl          time.Duration
	cacheEnabled bool
}

// NewAccountCacheDecorator wraps an AccountRepository with Redis caching.
func NewAccountCacheDecorator(
	repo domain.AccountRepository,
	redisClient *RedisClient,
	cfg *config.Config,
) domain.AccountRepository {
	return &AccountCacheDecorator{
		repo:         repo,
		redis:        redisClient,
		ttl:          cfg.CacheAccountTTL,
		cacheEnabled: cfg.CacheEnabled,
	}
}

// WithTx returns the underlying repo's WithTx — cache is bypassed inside DB transactions.
func (c *AccountCacheDecorator) WithTx(tx *gorm.DB) domain.AccountRepository {
	return c.repo.WithTx(tx)
}

// FetchByUserID checks cache first, falls back to PG on miss.
func (c *AccountCacheDecorator) FetchByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return c.repo.FetchByUserID(ctx, userID)
	}

	key := c.redis.BuildKey("account", "list", userID.String())

	var accounts []domain.Account
	err := c.redis.Get(ctx, key, &accounts)
	if err == nil {
		return accounts, nil
	}
	if err != redis.Nil {
		slog.ErrorContext(ctx, "cache get failed for account list", "key", key, "error", err)
	}

	accounts, err = c.repo.FetchByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if setErr := c.redis.Set(ctx, key, accounts, c.ttl); setErr != nil {
		slog.ErrorContext(ctx, "cache set failed for account list", "key", key, "error", setErr)
	}

	return accounts, nil
}

// GetByID checks cache first, falls back to PG on miss.
func (c *AccountCacheDecorator) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Account, error) {
	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return c.repo.GetByID(ctx, id, userID)
	}

	key := c.redis.BuildKey("account", userID.String(), id.String())

	var account domain.Account
	err := c.redis.Get(ctx, key, &account)
	if err == nil {
		return &account, nil
	}
	if err != redis.Nil {
		slog.ErrorContext(ctx, "cache get failed for account", "key", key, "error", err)
	}

	result, err := c.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if setErr := c.redis.Set(ctx, key, result, c.ttl); setErr != nil {
		slog.ErrorContext(ctx, "cache set failed for account", "key", key, "error", setErr)
	}

	return result, nil
}

// Create delegates to PG, then invalidates account + dashboard cache for the user.
func (c *AccountCacheDecorator) Create(ctx context.Context, account *domain.Account) error {
	if err := c.repo.Create(ctx, account); err != nil {
		return err
	}
	c.invalidateForUser(ctx, account.UserID)
	return nil
}

// Update delegates to PG, then invalidates account + dashboard cache for the user.
func (c *AccountCacheDecorator) Update(ctx context.Context, account *domain.Account) error {
	if err := c.repo.Update(ctx, account); err != nil {
		return err
	}
	c.invalidateForUser(ctx, account.UserID)
	return nil
}

// SoftDelete delegates to PG, then invalidates account + dashboard cache for the user.
func (c *AccountCacheDecorator) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	if err := c.repo.SoftDelete(ctx, id, userID); err != nil {
		return err
	}
	c.invalidateForUser(ctx, userID)
	return nil
}

// UpdateBalance delegates to PG, then invalidates account + dashboard cache for the user.
// Note: UpdateBalance only receives account ID and delta, not userID directly.
// We use DeleteByPattern with the account ID to invalidate the specific account key,
// and also invalidate the dashboard key pattern since balance changes affect summaries.
func (c *AccountCacheDecorator) UpdateBalance(ctx context.Context, id uuid.UUID, delta int64) error {
	if err := c.repo.UpdateBalance(ctx, id, delta); err != nil {
		return err
	}

	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return nil
	}

	// Invalidate any account key containing this account ID
	pattern := c.redis.BuildKey("account", "*"+id.String()+"*")
	if err := c.redis.DeleteByPattern(ctx, pattern); err != nil {
		slog.ErrorContext(ctx, "cache invalidation failed for account balance", "account_id", id, "error", err)
	}

	return nil
}

// GetAllActive delegates directly to PG (no caching).
func (c *AccountCacheDecorator) GetAllActive(ctx context.Context) ([]domain.Account, error) {
	return c.repo.GetAllActive(ctx)
}

// invalidateForUser deletes all account cache keys and the dashboard key for a user.
func (c *AccountCacheDecorator) invalidateForUser(ctx context.Context, userID uuid.UUID) {
	if !c.cacheEnabled || !c.redis.IsAvailable() {
		return
	}

	accountPattern := c.redis.BuildKey("account", "*"+userID.String()+"*")
	if err := c.redis.DeleteByPattern(ctx, accountPattern); err != nil {
		slog.ErrorContext(ctx, "cache invalidation failed for accounts", "user_id", userID, "error", err)
	}

	dashboardKey := c.redis.BuildKey("dashboard", userID.String())
	if err := c.redis.Delete(ctx, dashboardKey); err != nil {
		slog.ErrorContext(ctx, "cache invalidation failed for dashboard", "user_id", userID, "error", err)
	}
}
