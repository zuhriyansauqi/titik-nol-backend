package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "titik-nol"

// RedisClient wraps go-redis and provides typed get/set/delete with JSON serialization.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a Redis connection using config values.
// Returns a client in fallback mode (nil internal client) if connection fails.
func NewRedisClient(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Startup code — no request context available, slog without context is fine here.
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Warn("redis connection failed, operating in fallback mode", "error", err)
		_ = client.Close()
		return &RedisClient{client: nil}
	}

	slog.Info("redis connected", "addr", fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort))
	return &RedisClient{client: client}
}

// Ping verifies Redis connectivity.
func (r *RedisClient) Ping(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not available")
	}
	return r.client.Ping(ctx).Err()
}

// Get retrieves and deserializes a cached value into dest.
// Returns redis.Nil on cache miss.
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	if r.client == nil {
		return redis.Nil
	}
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set serializes and stores a value with the given TTL.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not available")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes one or more keys.
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not available")
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// DeleteByPattern removes all keys matching a glob pattern using SCAN.
func (r *RedisClient) DeleteByPattern(ctx context.Context, pattern string) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not available")
	}

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan keys: %w", err)
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("delete scanned keys: %w", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// IsAvailable returns true if the Redis client is connected and healthy.
func (r *RedisClient) IsAvailable() bool {
	return r.client != nil
}

// Close gracefully closes the Redis connection.
func (r *RedisClient) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}

// BuildKey constructs a cache key: "titik-nol:<parts joined by :>"
func (r *RedisClient) BuildKey(parts ...string) string {
	return keyPrefix + ":" + strings.Join(parts, ":")
}
