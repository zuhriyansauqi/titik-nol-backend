package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// setRequiredViperValues sets the mandatory config values so LoadConfig's validate() passes.
func setRequiredViperValues() {
	viper.Set("APP_PORT", "8080")
	viper.Set("DB_HOST", "localhost")
	viper.Set("DB_PORT", "5432")
	viper.Set("DB_USER", "admin")
	viper.Set("DB_NAME", "testdb")
	viper.Set("JWT_SECRET", "secret")
	viper.Set("GOOGLE_CLIENT_ID", "client-id")
}

// Feature: redis-caching, Property 4: Config loading reads all Redis and cache environment variables
func TestConfigLoad_RedisAndCacheEnvVars(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		viper.Reset()
		setRequiredViperValues()

		host := rapid.StringMatching(`[a-z][a-z0-9\-]{0,20}`).Draw(t, "host")
		port := rapid.IntRange(1024, 65535).Draw(t, "port")
		password := rapid.StringMatching(`[a-zA-Z0-9]{0,30}`).Draw(t, "password")
		db := rapid.IntRange(0, 15).Draw(t, "db")
		enabled := rapid.Bool().Draw(t, "enabled")

		// Generate TTL values in whole minutes (1–60 min)
		defaultTTLMin := rapid.IntRange(1, 60).Draw(t, "defaultTTLMin")
		categoryTTLMin := rapid.IntRange(1, 60).Draw(t, "categoryTTLMin")
		dashboardTTLMin := rapid.IntRange(1, 60).Draw(t, "dashboardTTLMin")
		userTTLMin := rapid.IntRange(1, 60).Draw(t, "userTTLMin")
		accountTTLMin := rapid.IntRange(1, 60).Draw(t, "accountTTLMin")

		viper.Set("REDIS_HOST", host)
		viper.Set("REDIS_PORT", fmt.Sprintf("%d", port))
		viper.Set("REDIS_PASSWORD", password)
		viper.Set("REDIS_DB", db)
		viper.Set("CACHE_ENABLED", enabled)
		viper.Set("CACHE_DEFAULT_TTL", time.Duration(defaultTTLMin)*time.Minute)
		viper.Set("CACHE_CATEGORY_TTL", time.Duration(categoryTTLMin)*time.Minute)
		viper.Set("CACHE_DASHBOARD_TTL", time.Duration(dashboardTTLMin)*time.Minute)
		viper.Set("CACHE_USER_TTL", time.Duration(userTTLMin)*time.Minute)
		viper.Set("CACHE_ACCOUNT_TTL", time.Duration(accountTTLMin)*time.Minute)

		cfg, err := LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, host, cfg.RedisHost)
		assert.Equal(t, fmt.Sprintf("%d", port), cfg.RedisPort)
		assert.Equal(t, password, cfg.RedisPassword)
		assert.Equal(t, db, cfg.RedisDB)
		assert.Equal(t, enabled, cfg.CacheEnabled)
		assert.Equal(t, time.Duration(defaultTTLMin)*time.Minute, cfg.CacheDefaultTTL)
		assert.Equal(t, time.Duration(categoryTTLMin)*time.Minute, cfg.CacheCategoryTTL)
		assert.Equal(t, time.Duration(dashboardTTLMin)*time.Minute, cfg.CacheDashboardTTL)
		assert.Equal(t, time.Duration(userTTLMin)*time.Minute, cfg.CacheUserTTL)
		assert.Equal(t, time.Duration(accountTTLMin)*time.Minute, cfg.CacheAccountTTL)
	})
}

func TestConfigDefaults_RedisCacheValues(t *testing.T) {
	viper.Reset()
	setRequiredViperValues()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.RedisHost)
	assert.Equal(t, "6379", cfg.RedisPort)
	assert.Equal(t, "", cfg.RedisPassword)
	assert.Equal(t, 0, cfg.RedisDB)
	assert.Equal(t, true, cfg.CacheEnabled)
	assert.Equal(t, 5*time.Minute, cfg.CacheDefaultTTL)
	assert.Equal(t, 30*time.Minute, cfg.CacheCategoryTTL)
	assert.Equal(t, 2*time.Minute, cfg.CacheDashboardTTL)
	assert.Equal(t, 10*time.Minute, cfg.CacheUserTTL)
	assert.Equal(t, 5*time.Minute, cfg.CacheAccountTTL)
}
