package domain

import (
	"context"

	"github.com/google/uuid"
)

// OnboardingUsecase defines the business logic interface for onboarding (titik nol setup).
type OnboardingUsecase interface {
	SetupAccounts(ctx context.Context, userID uuid.UUID, req *SetupAccountsRequest) (*SetupAccountsResponse, error)
}

// SetupAccountItem represents a single account in the bulk onboarding request.
type SetupAccountItem struct {
	Name           string      `json:"name" binding:"required"`
	Type           AccountType `json:"type" binding:"required"`
	InitialBalance int64       `json:"initial_balance" binding:"min=0"`
}

// SetupAccountsRequest is the DTO for bulk account setup during onboarding.
type SetupAccountsRequest struct {
	Accounts []SetupAccountItem `json:"accounts" binding:"required,min=1,dive"`
}

// SetupAccountsResponse is the DTO returned after successful onboarding setup.
type SetupAccountsResponse struct {
	Accounts     []Account     `json:"accounts"`
	Transactions []Transaction `json:"transactions"`
}
