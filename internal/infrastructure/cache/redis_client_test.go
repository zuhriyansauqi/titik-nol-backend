package cache

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// Feature: redis-caching, Property 2: Cache key determinism and format
func TestBuildKey_DeterministicFormat(t *testing.T) {
	rc := &RedisClient{}

	rapid.Check(t, func(t *rapid.T) {
		entity := rapid.StringMatching(`[a-z]{3,15}`).Draw(t, "entity")
		userID := genUUID(t, "user_id").String()
		qualifier := rapid.StringMatching(`[a-zA-Z0-9_]{0,20}`).Draw(t, "qualifier")

		parts := []string{entity, userID}
		if qualifier != "" {
			parts = append(parts, qualifier)
		}

		key1 := rc.BuildKey(parts...)
		key2 := rc.BuildKey(parts...)

		// Determinism: same inputs produce same key
		assert.Equal(t, key1, key2)

		// Format: must start with "titik-nol:" prefix
		assert.True(t, strings.HasPrefix(key1, "titik-nol:"))

		// Format: key equals "titik-nol:" + parts joined by ":"
		expected := "titik-nol:" + strings.Join(parts, ":")
		assert.Equal(t, expected, key1)
	})
}

// Verify specific key patterns from the design document
func TestBuildKey_CategoryListFormat(t *testing.T) {
	rc := &RedisClient{}

	rapid.Check(t, func(t *rapid.T) {
		userID := genUUID(t, "user_id").String()
		filterType := rapid.SampledFrom([]string{"INCOME", "EXPENSE", "all"}).Draw(t, "filter")

		key := rc.BuildKey("category", "list", userID, filterType)
		expected := "titik-nol:category:list:" + userID + ":" + filterType
		assert.Equal(t, expected, key)
	})
}

func TestBuildKey_AccountListFormat(t *testing.T) {
	rc := &RedisClient{}

	rapid.Check(t, func(t *rapid.T) {
		userID := genUUID(t, "user_id").String()

		key := rc.BuildKey("account", "list", userID)
		expected := "titik-nol:account:list:" + userID
		assert.Equal(t, expected, key)
	})
}

func TestBuildKey_DashboardFormat(t *testing.T) {
	rc := &RedisClient{}

	rapid.Check(t, func(t *rapid.T) {
		userID := genUUID(t, "user_id").String()

		key := rc.BuildKey("dashboard", userID)
		expected := "titik-nol:dashboard:" + userID
		assert.Equal(t, expected, key)
	})
}

// Feature: redis-caching, Property 3: Cache key collision freedom
func TestBuildKey_CollisionFreedom(t *testing.T) {
	rc := &RedisClient{}

	rapid.Check(t, func(t *rapid.T) {
		entity1 := rapid.StringMatching(`[a-z]{3,15}`).Draw(t, "entity1")
		userID1 := genUUID(t, "user_id1").String()
		qualifier1 := rapid.StringMatching(`[a-zA-Z0-9_]{0,20}`).Draw(t, "qualifier1")

		entity2 := rapid.StringMatching(`[a-z]{3,15}`).Draw(t, "entity2")
		userID2 := genUUID(t, "user_id2").String()
		qualifier2 := rapid.StringMatching(`[a-zA-Z0-9_]{0,20}`).Draw(t, "qualifier2")

		// Only test when tuples are actually distinct
		if entity1 == entity2 && userID1 == userID2 && qualifier1 == qualifier2 {
			return
		}

		parts1 := []string{entity1, userID1}
		if qualifier1 != "" {
			parts1 = append(parts1, qualifier1)
		}
		parts2 := []string{entity2, userID2}
		if qualifier2 != "" {
			parts2 = append(parts2, qualifier2)
		}

		key1 := rc.BuildKey(parts1...)
		key2 := rc.BuildKey(parts2...)

		assert.NotEqual(t, key1, key2, "distinct tuples must produce distinct keys")
	})
}
