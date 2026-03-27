package cache

import (
	"context"
	"testing"
	"time"

	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Feature: redis-caching, Property 7: User cache decorator read transparency
func TestUserCacheRepo_ReadTransparency(t *testing.T) {
	rc, _ := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		mockRepo := new(mocks.MockUserRepository)
		repo := &UserCacheDecorator{
			repo:         mockRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: true,
		}

		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		expectedUser := &domain.User{
			ID:         userID,
			Email:      rapid.StringMatching(`[a-z]{3,10}@[a-z]{3,8}\.com`).Draw(rt, "email"),
			Name:       rapid.StringMatching(`[a-zA-Z ]{1,30}`).Draw(rt, "name"),
			AvatarURL:  rapid.StringMatching(`https://example\.com/[a-z]{1,20}`).Draw(rt, "avatar"),
			Provider:   domain.ProviderGoogle,
			ProviderID: rapid.StringMatching(`[0-9]{10,20}`).Draw(rt, "provider_id"),
			Role:       domain.RoleUser,
			CreatedAt:  genTime(rt, "created"),
			UpdatedAt:  genTime(rt, "updated"),
		}
		mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil).Once()

		// First call: cache miss, hits mock repo
		got, err := repo.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, got)

		// Second call: cache hit, must NOT call mock repo again
		got2, err := repo.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, got2)

		mockRepo.AssertExpectations(t)
	})
}

// Feature: redis-caching, Property 11: User write invalidation
func TestUserCacheRepo_WriteInvalidation(t *testing.T) {
	rc, mr := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		mockRepo := new(mocks.MockUserRepository)
		repo := &UserCacheDecorator{
			repo:         mockRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: true,
		}

		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		// Pre-populate cache with a user entry
		key := rc.BuildKey("user", userID.String())
		require.NoError(t, rc.Set(ctx, key, domain.User{ID: userID}, 5*time.Minute))
		assert.True(t, mr.Exists(key))

		// Update the user
		user := &domain.User{
			ID:         userID,
			Email:      rapid.StringMatching(`[a-z]{3,10}@[a-z]{3,8}\.com`).Draw(rt, "email"),
			Name:       rapid.StringMatching(`[a-zA-Z ]{1,30}`).Draw(rt, "name"),
			AvatarURL:  rapid.StringMatching(`https://example\.com/[a-z]{1,20}`).Draw(rt, "avatar"),
			Provider:   domain.ProviderGoogle,
			ProviderID: rapid.StringMatching(`[0-9]{10,20}`).Draw(rt, "provider_id"),
			Role:       domain.RoleUser,
			CreatedAt:  genTime(rt, "created"),
			UpdatedAt:  genTime(rt, "updated"),
		}
		mockRepo.On("Update", ctx, user).Return(nil).Once()

		err := repo.Update(ctx, user)
		require.NoError(t, err)

		// Cached user entry must be absent after Update
		assert.False(t, mr.Exists(key), "user cache key should be invalidated after Update")

		mockRepo.AssertExpectations(t)
	})
}
