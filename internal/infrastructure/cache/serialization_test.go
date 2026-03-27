package cache

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// --- Generators ---

func genUUID(t *rapid.T, label string) uuid.UUID {
	b := rapid.SliceOfN(rapid.Byte(), 16, 16).Draw(t, label)
	id, _ := uuid.FromBytes(b)
	return id
}

func genTime(t *rapid.T, label string) time.Time {
	sec := rapid.Int64Range(0, 2000000000).Draw(t, label)
	return time.Unix(sec, 0).UTC()
}

func genTimePtr(t *rapid.T, label string) *time.Time {
	if rapid.Bool().Draw(t, label+"_nil") {
		return nil
	}
	ts := genTime(t, label)
	return &ts
}

func genUUIDPtr(t *rapid.T, label string) *uuid.UUID {
	if rapid.Bool().Draw(t, label+"_nil") {
		return nil
	}
	id := genUUID(t, label)
	return &id
}

func genAccountType(t *rapid.T, label string) domain.AccountType {
	types := []domain.AccountType{
		domain.AccountTypeCash, domain.AccountTypeBank,
		domain.AccountTypeEWallet, domain.AccountTypeCreditCard,
	}
	return types[rapid.IntRange(0, len(types)-1).Draw(t, label)]
}

func genCategoryType(t *rapid.T, label string) domain.CategoryType {
	types := []domain.CategoryType{domain.CategoryTypeIncome, domain.CategoryTypeExpense}
	return types[rapid.IntRange(0, len(types)-1).Draw(t, label)]
}

func genTxType(t *rapid.T, label string) domain.TransactionType {
	types := []domain.TransactionType{
		domain.TxTypeIncome, domain.TxTypeExpense,
		domain.TxTypeTransfer, domain.TxTypeAdjustment,
	}
	return types[rapid.IntRange(0, len(types)-1).Draw(t, label)]
}

func genAccount(t *rapid.T) domain.Account {
	return domain.Account{
		ID:        genUUID(t, "acc_id"),
		UserID:    genUUID(t, "acc_user_id"),
		Name:      rapid.StringMatching(`[a-zA-Z0-9 ]{1,50}`).Draw(t, "acc_name"),
		Type:      genAccountType(t, "acc_type"),
		Balance:   rapid.Int64Range(-1000000, 1000000).Draw(t, "acc_balance"),
		CreatedAt: genTime(t, "acc_created"),
		UpdatedAt: genTime(t, "acc_updated"),
		DeletedAt: genTimePtr(t, "acc_deleted"),
	}
}

func genCategory(t *rapid.T) domain.Category {
	return domain.Category{
		ID:        genUUID(t, "cat_id"),
		UserID:    genUUID(t, "cat_user_id"),
		Name:      rapid.StringMatching(`[a-zA-Z0-9 ]{1,50}`).Draw(t, "cat_name"),
		Type:      genCategoryType(t, "cat_type"),
		Icon:      rapid.StringMatching(`[a-z]{0,20}`).Draw(t, "cat_icon"),
		CreatedAt: genTime(t, "cat_created"),
	}
}

func genUser(t *rapid.T) domain.User {
	providers := []domain.AuthProvider{domain.ProviderGoogle, domain.ProviderLocal}
	roles := []domain.UserRole{domain.RoleUser, domain.RoleAdmin}

	var password *string
	if rapid.Bool().Draw(t, "user_has_password") {
		p := rapid.StringMatching(`[a-zA-Z0-9]{8,30}`).Draw(t, "user_password")
		password = &p
	}

	return domain.User{
		ID:         genUUID(t, "user_id"),
		Email:      rapid.StringMatching(`[a-z]{3,10}@[a-z]{3,8}\.com`).Draw(t, "user_email"),
		Name:       rapid.StringMatching(`[a-zA-Z ]{1,50}`).Draw(t, "user_name"),
		AvatarURL:  rapid.StringMatching(`https://example\.com/[a-z]{1,20}`).Draw(t, "user_avatar"),
		Provider:   providers[rapid.IntRange(0, len(providers)-1).Draw(t, "user_provider")],
		ProviderID: rapid.StringMatching(`[0-9]{10,20}`).Draw(t, "user_provider_id"),
		Password:   password,
		Role:       roles[rapid.IntRange(0, len(roles)-1).Draw(t, "user_role")],
		CreatedAt:  genTime(t, "user_created"),
		UpdatedAt:  genTime(t, "user_updated"),
	}
}

func genTransaction(t *rapid.T, label string) domain.Transaction {
	return domain.Transaction{
		ID:              genUUID(t, label+"_id"),
		UserID:          genUUID(t, label+"_user_id"),
		AccountID:       genUUID(t, label+"_account_id"),
		CategoryID:      genUUIDPtr(t, label+"_category_id"),
		TransactionType: genTxType(t, label+"_type"),
		Amount:          rapid.Int64Range(1, 1000000).Draw(t, label+"_amount"),
		Note:            rapid.StringMatching(`[a-zA-Z0-9 ]{0,50}`).Draw(t, label+"_note"),
		TransactionDate: genTime(t, label+"_date"),
		CreatedAt:       genTime(t, label+"_created"),
		DeletedAt:       genTimePtr(t, label+"_deleted"),
	}
}

func genDashboardSummary(t *rapid.T) domain.DashboardSummary {
	txCount := rapid.IntRange(0, 5).Draw(t, "tx_count")
	txs := make([]domain.Transaction, txCount)
	for i := range txs {
		txs[i] = genTransaction(t, fmt.Sprintf("tx_%d", i))
	}
	return domain.DashboardSummary{
		TotalBalance:       rapid.Int64Range(-10000000, 10000000).Draw(t, "total_balance"),
		RecentTransactions: txs,
		NeedsPaydaySetup:   rapid.Bool().Draw(t, "needs_payday"),
	}
}

// Feature: redis-caching, Property 1: Serialization round-trip
func TestSerialization_RoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Account round-trip
		origAccount := genAccount(t)
		data, err := json.Marshal(origAccount)
		require.NoError(t, err)
		var gotAccount domain.Account
		require.NoError(t, json.Unmarshal(data, &gotAccount))
		assert.Equal(t, origAccount, gotAccount)

		// Category round-trip
		origCategory := genCategory(t)
		data, err = json.Marshal(origCategory)
		require.NoError(t, err)
		var gotCategory domain.Category
		require.NoError(t, json.Unmarshal(data, &gotCategory))
		assert.Equal(t, origCategory, gotCategory)

		// User round-trip
		// Note: Password field has json:"-" tag, so it is excluded from serialization.
		// After unmarshal, Password will always be nil regardless of original value.
		origUser := genUser(t)
		data, err = json.Marshal(origUser)
		require.NoError(t, err)
		var gotUser domain.User
		require.NoError(t, json.Unmarshal(data, &gotUser))
		expectedUser := origUser
		expectedUser.Password = nil
		assert.Equal(t, expectedUser, gotUser)

		// DashboardSummary round-trip
		origDash := genDashboardSummary(t)
		data, err = json.Marshal(origDash)
		require.NoError(t, err)
		var gotDash domain.DashboardSummary
		require.NoError(t, json.Unmarshal(data, &gotDash))
		// Handle nil vs empty slice for RecentTransactions
		if len(origDash.RecentTransactions) == 0 && len(gotDash.RecentTransactions) == 0 {
			origDash.RecentTransactions = nil
			gotDash.RecentTransactions = nil
		}
		assert.Equal(t, origDash, gotDash)
	})
}
