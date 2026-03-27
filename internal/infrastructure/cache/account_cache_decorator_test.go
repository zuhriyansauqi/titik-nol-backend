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

func genAccountSlice(t *rapid.T) []domain.Account {
	count := rapid.IntRange(0, 5).Draw(t, "acc_count")
	accs := make([]domain.Account, count)
	for i := range accs {
		accs[i] = genAccount(t)
		if i > 0 {
			accs[i].UserID = accs[0].UserID
		}
	}
	return accs
}

// Feature: redis-caching, Property 6: Account cache decorator read transparency
func TestAccountCacheRepo_ReadTransparency(t *testing.T) {
	rc, _ := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		mockRepo := new(mocks.MockAccountRepository)
		repo := &AccountCacheDecorator{
			repo:         mockRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: true,
		}

		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		// --- FetchByUserID transparency ---
		expectedList := genAccountSlice(rt)
		mockRepo.On("FetchByUserID", ctx, userID).Return(expectedList, nil).Once()

		// First call: cache miss, hits mock repo
		gotList, err := repo.FetchByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedList, gotList)

		// Second call: cache hit, must NOT call mock repo again
		gotList2, err := repo.FetchByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedList, gotList2)

		// --- GetByID transparency ---
		accID := genUUID(rt, "acc_id")
		expectedAcc := &domain.Account{
			ID:        accID,
			UserID:    userID,
			Name:      rapid.StringMatching(`[a-zA-Z0-9 ]{1,30}`).Draw(rt, "acc_name"),
			Type:      genAccountType(rt, "acc_type"),
			Balance:   rapid.Int64Range(-1000000, 1000000).Draw(rt, "acc_balance"),
			CreatedAt: genTime(rt, "acc_created"),
			UpdatedAt: genTime(rt, "acc_updated"),
			DeletedAt: genTimePtr(rt, "acc_deleted"),
		}
		mockRepo.On("GetByID", ctx, accID, userID).Return(expectedAcc, nil).Once()

		gotAcc, err := repo.GetByID(ctx, accID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAcc, gotAcc)

		gotAcc2, err := repo.GetByID(ctx, accID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedAcc, gotAcc2)

		mockRepo.AssertExpectations(t)
	})
}

// Feature: redis-caching, Property 10: Account write invalidation with cross-entity dashboard invalidation
func TestAccountCacheRepo_WriteInvalidatesAccountAndDashboard(t *testing.T) {
	rc, mr := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		mockRepo := new(mocks.MockAccountRepository)
		repo := &AccountCacheDecorator{
			repo:         mockRepo,
			redis:        rc,
			ttl:          5 * time.Minute,
			cacheEnabled: true,
		}

		ctx := context.Background()
		userID := genUUID(rt, "user_id")
		accID := genUUID(rt, "acc_id")

		// Pre-populate cache with account list, individual entry, and dashboard
		listKey := rc.BuildKey("account", "list", userID.String())
		itemKey := rc.BuildKey("account", userID.String(), accID.String())
		dashKey := rc.BuildKey("dashboard", userID.String())

		require.NoError(t, rc.Set(ctx, listKey, []domain.Account{}, 5*time.Minute))
		require.NoError(t, rc.Set(ctx, itemKey, domain.Account{}, 5*time.Minute))
		require.NoError(t, rc.Set(ctx, dashKey, domain.DashboardSummary{}, 5*time.Minute))

		assert.True(t, mr.Exists(listKey))
		assert.True(t, mr.Exists(itemKey))
		assert.True(t, mr.Exists(dashKey))

		// Pick a random write operation
		op := rapid.IntRange(0, 2).Draw(rt, "write_op")
		switch op {
		case 0: // Create
			acc := &domain.Account{
				ID:        accID,
				UserID:    userID,
				Name:      rapid.StringMatching(`[a-zA-Z0-9 ]{1,30}`).Draw(rt, "acc_name"),
				Type:      genAccountType(rt, "acc_type"),
				Balance:   rapid.Int64Range(0, 1000000).Draw(rt, "acc_balance"),
				CreatedAt: genTime(rt, "acc_created"),
				UpdatedAt: genTime(rt, "acc_updated"),
			}
			mockRepo.On("Create", ctx, acc).Return(nil).Once()
			require.NoError(t, repo.Create(ctx, acc))
		case 1: // Update
			acc := &domain.Account{
				ID:        accID,
				UserID:    userID,
				Name:      rapid.StringMatching(`[a-zA-Z0-9 ]{1,30}`).Draw(rt, "acc_name"),
				Type:      genAccountType(rt, "acc_type"),
				Balance:   rapid.Int64Range(0, 1000000).Draw(rt, "acc_balance"),
				CreatedAt: genTime(rt, "acc_created"),
				UpdatedAt: genTime(rt, "acc_updated"),
			}
			mockRepo.On("Update", ctx, acc).Return(nil).Once()
			require.NoError(t, repo.Update(ctx, acc))
		case 2: // SoftDelete
			mockRepo.On("SoftDelete", ctx, accID, userID).Return(nil).Once()
			require.NoError(t, repo.SoftDelete(ctx, accID, userID))
		}

		// All account cache entries and dashboard entry for this user must be absent
		assert.False(t, mr.Exists(listKey), "account list key should be invalidated")
		assert.False(t, mr.Exists(itemKey), "individual account key should be invalidated")
		assert.False(t, mr.Exists(dashKey), "dashboard key should be invalidated")

		mockRepo.AssertExpectations(t)
	})
}
