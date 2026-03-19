package domain

import (
	"context"

	"github.com/google/uuid"
)

type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	IsNewUser   bool   `json:"is_new_user"`
}

type AuthUsecase interface {
	LoginWithGoogle(ctx context.Context, req *GoogleLoginRequest) (*AuthResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*User, error)
}
