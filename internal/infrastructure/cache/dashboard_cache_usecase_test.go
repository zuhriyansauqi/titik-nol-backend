package cache

import (
	"context"
	"testing"
	"time"

	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Feature: redis-caching, Property 8: Dashboard cache decorator read transparency
func TestDashboardCacheUsecase_ReadTransparency(t *testing.T) {
	rc, _ := newTestRedisClient(t)

	rapid.Check(t, func(rt *rapid.T) {
		_ = rc.client.FlushAll(context.Background())

		mockUC := new(mocks.MockDashboardUsecase)
		uc := &DashboardCacheUsecase{
			usecase:      mockUC,
			redis:        rc,
			ttl:          2 * time.Minute,
			cacheEnabled: true,
		}

		ctx := context.Background()
		userID := genUUID(rt, "user_id")

		expected := genDashboardSummary(rt)
		mockUC.On("GetSummary", ctx, userID).Return(&expected, nil).Once()

		// First call: cache miss, hits mock usecase
		got, err := uc.GetSummary(ctx, userID)
		require.NoError(t, err)
		// Handle nil vs empty slice for RecentTransactions
		if len(expected.RecentTransactions) == 0 && len(got.RecentTransactions) == 0 {
			expected.RecentTransactions = nil
			got.RecentTransactions = nil
		}
		assert.Equal(t, &expected, got)

		// Second call: cache hit, must NOT call mock usecase again
		got2, err := uc.GetSummary(ctx, userID)
		require.NoError(t, err)
		if len(got2.RecentTransactions) == 0 {
			got2.RecentTransactions = nil
		}
		assert.Equal(t, &expected, got2)

		mockUC.AssertExpectations(t)
	})
}
