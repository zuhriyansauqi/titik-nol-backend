package domain

import (
	"context"

	"github.com/google/uuid"
)

// DashboardUsecase defines the business logic interface for the dashboard summary.
type DashboardUsecase interface {
	GetSummary(ctx context.Context, userID uuid.UUID) (*DashboardSummary, error)
}

// DashboardSummary is the DTO returned by the dashboard summary endpoint.
type DashboardSummary struct {
	TotalBalance       int64         `json:"total_balance"`
	RecentTransactions []Transaction `json:"recent_transactions"`
	NeedsPaydaySetup   bool          `json:"needs_payday_setup"`
}
