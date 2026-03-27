package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// newTestRedisClient creates a RedisClient backed by miniredis for testing.
func newTestRedisClient(t *testing.T) (*RedisClient, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return &RedisClient{client: client}, mr
}

func genCategorySlice(t *rapid.T) []domain.Category {
	count := rapid.IntRange(0, 5).Draw(t, "cat_count")
	cats := make([]domain.Category, count)
	for i := range cats {
		cats[i] = genCategory(t)
		if i > 0 {
			cats[i].UserID = cats[0].UserID
		}
	}
	return cats
}

func genFilterType(t *rapid.T) *domain.CategoryType {
	if rapid.Bool().Draw(t, "has_filter") {
		ct := genCategoryType(t, "filter_type")
		return &ct
	}
	return nil
}

// Feature: redis-caching, Property 5: Category cache decorator read transparency
func TestCategoryCacheRepo_ReadTransparency(t *testing.T) {
	rc, _ := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		// Flush Redis between iterations to avoid cross-iteration cache hits
		_ = rc.client.FlushAll(context.Background())

		mockRepo := new(mocks.MockCategoryRepository)
		repo := &CategoryCacheDecorator{
			repo:         mockRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: true,
		}

		ctx := context.Background()
		userID := genUUID(rt, "user_id")
		filterType := genFilterType(rt)

		// --- FetchByUserID transparency ---
		expectedList := genCategorySlice(rt)
		mockRepo.On("FetchByUserID", ctx, userID, filterType).Return(expectedList, nil).Once()

		// First call: cache miss, hits mock repo
		gotList, err := repo.FetchByUserID(ctx, userID, filterType)
		require.NoError(t, err)
		assert.Equal(t, expectedList, gotList)

		// Second call: cache hit, must NOT call mock repo again
		gotList2, err := repo.FetchByUserID(ctx, userID, filterType)
		require.NoError(t, err)
		assert.Equal(t, expectedList, gotList2)

		// --- GetByID transparency ---
		catID := genUUID(rt, "cat_id")
		expectedCat := &domain.Category{
			ID:        catID,
			UserID:    userID,
			Name:      rapid.StringMatching(`[a-zA-Z0-9 ]{1,30}`).Draw(rt, "cat_name"),
			Type:      genCategoryType(rt, "cat_type"),
			Icon:      rapid.StringMatching(`[a-z]{0,10}`).Draw(rt, "cat_icon"),
			CreatedAt: genTime(rt, "cat_created"),
		}
		mockRepo.On("GetByID", ctx, catID, userID).Return(expectedCat, nil).Once()

		gotCat, err := repo.GetByID(ctx, catID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCat, gotCat)

		gotCat2, err := repo.GetByID(ctx, catID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCat, gotCat2)

		mockRepo.AssertExpectations(t)
	})
}

// Feature: redis-caching, Property 9: Category write invalidation
func TestCategoryCacheRepo_WriteInvalidation(t *testing.T) {
	rc, mr := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		mockRepo := new(mocks.MockCategoryRepository)
		repo := &CategoryCacheDecorator{
			repo:         mockRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: true,
		}

		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		// Pre-populate cache with category list and individual entries
		listKey := rc.BuildKey("category", "list", userID.String(), "all")
		incomeKey := rc.BuildKey("category", "list", userID.String(), "income")
		catID := genUUID(rt, "cat_id")
		itemKey := rc.BuildKey("category", userID.String(), catID.String())

		require.NoError(t, rc.Set(ctx, listKey, []domain.Category{}, 5*time.Minute))
		require.NoError(t, rc.Set(ctx, incomeKey, []domain.Category{}, 5*time.Minute))
		require.NoError(t, rc.Set(ctx, itemKey, domain.Category{}, 5*time.Minute))

		// Verify keys exist before Create
		assert.True(t, mr.Exists(listKey))
		assert.True(t, mr.Exists(incomeKey))
		assert.True(t, mr.Exists(itemKey))

		// Create a new category
		newCat := &domain.Category{
			ID:        genUUID(rt, "new_cat_id"),
			UserID:    userID,
			Name:      rapid.StringMatching(`[a-zA-Z0-9 ]{1,30}`).Draw(rt, "new_cat_name"),
			Type:      genCategoryType(rt, "new_cat_type"),
			Icon:      rapid.StringMatching(`[a-z]{0,10}`).Draw(rt, "new_cat_icon"),
			CreatedAt: genTime(rt, "new_cat_created"),
		}
		mockRepo.On("Create", ctx, newCat).Return(nil).Once()

		err := repo.Create(ctx, newCat)
		require.NoError(t, err)

		// All category cache entries for this user must be absent
		assert.False(t, mr.Exists(listKey), "list:all key should be invalidated")
		assert.False(t, mr.Exists(incomeKey), "list:income key should be invalidated")
		assert.False(t, mr.Exists(itemKey), "individual category key should be invalidated")

		mockRepo.AssertExpectations(t)
	})
}
