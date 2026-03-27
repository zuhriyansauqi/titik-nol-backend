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

// Feature: redis-caching, Property 12: Cache-disabled bypass
func TestCacheDecorators_DisabledBypass(t *testing.T) {
	rc, mr := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		// --- Category decorator with cache disabled ---
		mockCatRepo := new(mocks.MockCategoryRepository)
		catDecorator := &CategoryCacheDecorator{
			repo:         mockCatRepo,
			redis:        rc,
			ttl:          30 * time.Minute,
			cacheEnabled: false,
		}

		expectedCats := genCategorySlice(rt)
		filterType := genFilterType(rt)
		mockCatRepo.On("FetchByUserID", ctx, userID, filterType).Return(expectedCats, nil).Twice()

		got, err := catDecorator.FetchByUserID(ctx, userID, filterType)
		require.NoError(t, err)
		assert.Equal(t, expectedCats, got)

		// Second call must also hit the repo (no caching)
		got2, err := catDecorator.FetchByUserID(ctx, userID, filterType)
		require.NoError(t, err)
		assert.Equal(t, expectedCats, got2)

		catID := genUUID(rt, "cat_id")
		expectedCat := &domain.Category{ID: catID, UserID: userID, Name: "test", Type: domain.CategoryTypeIncome}
		mockCatRepo.On("GetByID", ctx, catID, userID).Return(expectedCat, nil).Twice()

		gotCat, err := catDecorator.GetByID(ctx, catID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCat, gotCat)

		gotCat2, err := catDecorator.GetByID(ctx, catID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCat, gotCat2)

		// --- Account decorator with cache disabled ---
		mockAccRepo := new(mocks.MockAccountRepository)
		accDecorator := &AccountCacheDecorator{
			repo:         mockAccRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: false,
		}

		expectedAccs := genAccountSlice(rt)
		mockAccRepo.On("FetchByUserID", ctx, userID).Return(expectedAccs, nil).Twice()

		gotAccs, err := accDecorator.FetchByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAccs, gotAccs)

		gotAccs2, err := accDecorator.FetchByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAccs, gotAccs2)

		accID := genUUID(rt, "acc_id")
		expectedAcc := &domain.Account{ID: accID, UserID: userID, Name: "test", Type: domain.AccountTypeCash}
		mockAccRepo.On("GetByID", ctx, accID, userID).Return(expectedAcc, nil).Twice()

		gotAcc, err := accDecorator.GetByID(ctx, accID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAcc, gotAcc)

		gotAcc2, err := accDecorator.GetByID(ctx, accID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAcc, gotAcc2)

		// --- User decorator with cache disabled ---
		mockUserRepo := new(mocks.MockUserRepository)
		userDecorator := &UserCacheDecorator{
			repo:         mockUserRepo,
			redis:        rc,
			ttl:          10 * time.Minute,
			cacheEnabled: false,
		}

		expectedUser := &domain.User{ID: userID, Email: "test@test.com", Name: "Test", Provider: domain.ProviderGoogle, ProviderID: "123", Role: domain.RoleUser}
		mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil).Twice()

		gotUser, err := userDecorator.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, gotUser)

		gotUser2, err := userDecorator.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, gotUser2)

		// --- Dashboard decorator with cache disabled ---
		mockDashUC := new(mocks.MockDashboardUsecase)
		dashDecorator := &DashboardCacheUsecase{
			usecase:      mockDashUC,
			redis:        rc,
			ttl:          2 * time.Minute,
			cacheEnabled: false,
		}

		expectedDash := &domain.DashboardSummary{TotalBalance: 1000, NeedsPaydaySetup: false}
		mockDashUC.On("GetSummary", ctx, userID).Return(expectedDash, nil).Twice()

		gotDash, err := dashDecorator.GetSummary(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedDash, gotDash)

		gotDash2, err := dashDecorator.GetSummary(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedDash, gotDash2)

		// Verify no keys were written to Redis
		keys := mr.Keys()
		assert.Empty(t, keys, "no keys should be written to Redis when cache is disabled")

		mockCatRepo.AssertExpectations(t)
		mockAccRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockDashUC.AssertExpectations(t)
	})
}

// Feature: redis-caching, Property 13: Non-cached method bypass
func TestCacheDecorators_NonCachedMethodBypass(t *testing.T) {
	rc, mr := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		// --- CategoryCacheDecorator.CountByUserID bypasses Redis ---
		mockCatRepo := new(mocks.MockCategoryRepository)
		catDecorator := &CategoryCacheDecorator{
			repo:         mockCatRepo,
			redis:        rc,
			ttl:          30 * time.Minute,
			cacheEnabled: true,
		}

		expectedCount := rapid.Int64Range(0, 100).Draw(rt, "cat_count")
		mockCatRepo.On("CountByUserID", ctx, userID).Return(expectedCount, nil).Twice()

		gotCount, err := catDecorator.CountByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCount, gotCount)

		// Second call must also hit the repo (no caching)
		gotCount2, err := catDecorator.CountByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCount, gotCount2)

		// --- UserCacheDecorator.GetByEmail bypasses Redis ---
		mockUserRepo := new(mocks.MockUserRepository)
		userDecorator := &UserCacheDecorator{
			repo:         mockUserRepo,
			redis:        rc,
			ttl:          10 * time.Minute,
			cacheEnabled: true,
		}

		email := rapid.StringMatching(`[a-z]{3,10}@[a-z]{3,8}\.com`).Draw(rt, "email")
		expectedUser := &domain.User{ID: userID, Email: email, Name: "Test", Provider: domain.ProviderGoogle, ProviderID: "123", Role: domain.RoleUser}
		mockUserRepo.On("GetByEmail", ctx, email).Return(expectedUser, nil).Twice()

		gotUser, err := userDecorator.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, gotUser)

		gotUser2, err := userDecorator.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, gotUser2)

		// --- UserCacheDecorator.GetByProviderID bypasses Redis ---
		providerID := rapid.StringMatching(`[0-9]{10,20}`).Draw(rt, "provider_id")
		expectedUser2 := &domain.User{ID: userID, Email: email, Name: "Test", Provider: domain.ProviderGoogle, ProviderID: providerID, Role: domain.RoleUser}
		mockUserRepo.On("GetByProviderID", ctx, providerID).Return(expectedUser2, nil).Twice()

		gotUser3, err := userDecorator.GetByProviderID(ctx, providerID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser2, gotUser3)

		gotUser4, err := userDecorator.GetByProviderID(ctx, providerID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser2, gotUser4)

		// Verify no keys were written to Redis for these non-cached methods
		keys := mr.Keys()
		assert.Empty(t, keys, "no keys should be written to Redis for non-cached methods")

		mockCatRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

// newUnavailableRedisClient creates a RedisClient with nil internal client (simulates Redis down).
func newUnavailableRedisClient() *RedisClient {
	return &RedisClient{client: nil}
}

// Feature: redis-caching, Property 14: Graceful degradation on Redis failure
func TestCacheDecorators_GracefulDegradation(t *testing.T) {
	// Use a nil-client RedisClient to simulate Redis being completely unavailable.
	// IsAvailable() returns false, so all decorators fall back to the underlying repo/usecase.
	rc := newUnavailableRedisClient()

	rapid.Check(t, func(rt *rapid.T) {
		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		// --- Category decorator graceful degradation ---
		mockCatRepo := new(mocks.MockCategoryRepository)
		catDecorator := &CategoryCacheDecorator{
			repo:         mockCatRepo,
			redis:        rc,
			ttl:          30 * time.Minute,
			cacheEnabled: true,
		}

		expectedCats := genCategorySlice(rt)
		filterType := genFilterType(rt)
		mockCatRepo.On("FetchByUserID", ctx, userID, filterType).Return(expectedCats, nil).Once()

		gotCats, err := catDecorator.FetchByUserID(ctx, userID, filterType)
		require.NoError(t, err)
		assert.Equal(t, expectedCats, gotCats)

		catID := genUUID(rt, "cat_id")
		expectedCat := &domain.Category{ID: catID, UserID: userID, Name: "test", Type: domain.CategoryTypeIncome}
		mockCatRepo.On("GetByID", ctx, catID, userID).Return(expectedCat, nil).Once()

		gotCat, err := catDecorator.GetByID(ctx, catID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedCat, gotCat)

		// --- Account decorator graceful degradation ---
		mockAccRepo := new(mocks.MockAccountRepository)
		accDecorator := &AccountCacheDecorator{
			repo:         mockAccRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: true,
		}

		expectedAccs := genAccountSlice(rt)
		mockAccRepo.On("FetchByUserID", ctx, userID).Return(expectedAccs, nil).Once()

		gotAccs, err := accDecorator.FetchByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAccs, gotAccs)

		accID := genUUID(rt, "acc_id")
		expectedAcc := &domain.Account{ID: accID, UserID: userID, Name: "test", Type: domain.AccountTypeCash}
		mockAccRepo.On("GetByID", ctx, accID, userID).Return(expectedAcc, nil).Once()

		gotAcc, err := accDecorator.GetByID(ctx, accID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAcc, gotAcc)

		// Write operations should also succeed without error when Redis is down
		newAcc := &domain.Account{ID: accID, UserID: userID, Name: "new", Type: domain.AccountTypeCash}
		mockAccRepo.On("Create", ctx, newAcc).Return(nil).Once()
		require.NoError(t, accDecorator.Create(ctx, newAcc))

		mockAccRepo.On("Update", ctx, newAcc).Return(nil).Once()
		require.NoError(t, accDecorator.Update(ctx, newAcc))

		mockAccRepo.On("SoftDelete", ctx, accID, userID).Return(nil).Once()
		require.NoError(t, accDecorator.SoftDelete(ctx, accID, userID))

		// --- User decorator graceful degradation ---
		mockUserRepo := new(mocks.MockUserRepository)
		userDecorator := &UserCacheDecorator{
			repo:         mockUserRepo,
			redis:        rc,
			ttl:          10 * time.Minute,
			cacheEnabled: true,
		}

		expectedUser := &domain.User{ID: userID, Email: "test@test.com", Name: "Test", Provider: domain.ProviderGoogle, ProviderID: "123", Role: domain.RoleUser}
		mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil).Once()

		gotUser, err := userDecorator.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, gotUser)

		mockUserRepo.On("Update", ctx, expectedUser).Return(nil).Once()
		require.NoError(t, userDecorator.Update(ctx, expectedUser))

		// --- Dashboard decorator graceful degradation ---
		mockDashUC := new(mocks.MockDashboardUsecase)
		dashDecorator := &DashboardCacheUsecase{
			usecase:      mockDashUC,
			redis:        rc,
			ttl:          2 * time.Minute,
			cacheEnabled: true,
		}

		expectedDash := &domain.DashboardSummary{TotalBalance: rapid.Int64Range(-10000000, 10000000).Draw(rt, "balance"), NeedsPaydaySetup: rapid.Bool().Draw(rt, "payday")}
		mockDashUC.On("GetSummary", ctx, userID).Return(expectedDash, nil).Once()

		gotDash, err := dashDecorator.GetSummary(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedDash, gotDash)

		mockCatRepo.AssertExpectations(t)
		mockAccRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockDashUC.AssertExpectations(t)
	})
}
