package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/google"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/jwt"
)

type authUsecase struct {
	userRepo   domain.UserRepository
	googleSSO  google.GoogleSSOService
	jwtService jwt.JWTService
}

func NewAuthUsecase(userRepo domain.UserRepository, googleSSO google.GoogleSSOService, jwtService jwt.JWTService) domain.AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		googleSSO:  googleSSO,
		jwtService: jwtService,
	}
}

func (u *authUsecase) LoginWithGoogle(ctx context.Context, req *domain.GoogleLoginRequest) (*domain.AuthResponse, error) {
	slog.InfoContext(ctx, "Google login initiated")

	payload, err := u.googleSSO.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to verify Google ID Token", "error", err)
		return nil, err
	}

	providerID := payload.Subject

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		slog.WarnContext(ctx, "Google token missing 'email' claim")
		return nil, fmt.Errorf("google token missing required claim: email")
	}
	name, ok := payload.Claims["name"].(string)
	if !ok || name == "" {
		slog.WarnContext(ctx, "Google token missing 'name' claim")
		return nil, fmt.Errorf("google token missing required claim: name")
	}
	avatarURL, _ := payload.Claims["picture"].(string) // optional, no error if missing

	slog.InfoContext(ctx, "Google token verified", "email", email, "provider_id", providerID)

	user, err := u.userRepo.GetByProviderID(ctx, providerID)
	isNewUser := false

	if err != nil {
		slog.InfoContext(ctx, "User not found by ProviderID, checking by email", "email", email)
		// If user not found by ProviderID, check by email
		user, err = u.userRepo.GetByEmail(ctx, email)
		if err != nil {
			slog.InfoContext(ctx, "User not found by email, creating new user", "email", email)
			// Create new user (Auto-registration)
			user = &domain.User{
				Name:       name,
				Email:      email,
				Provider:   domain.ProviderGoogle,
				ProviderID: providerID,
				AvatarURL:  avatarURL,
			}
			if err := u.userRepo.Create(ctx, user); err != nil {
				slog.ErrorContext(ctx, "Failed to create user", "email", email, "error", err)
				return nil, err
			}
			isNewUser = true
			slog.InfoContext(ctx, "New user registered via Google SSO", "user_id", user.ID, "email", email)
		} else {
			slog.InfoContext(ctx, "User found by email, linking ProviderID", "email", email)
			// Link ProviderID to existing email
			user.ProviderID = providerID
			user.Provider = domain.ProviderGoogle
			user.AvatarURL = avatarURL
			if err := u.userRepo.Update(ctx, user); err != nil {
				slog.ErrorContext(ctx, "Failed to update user with ProviderID", "user_id", user.ID, "error", err)
				return nil, err
			}
			slog.InfoContext(ctx, "ProviderID linked to existing user", "user_id", user.ID, "email", email)
		}
	}

	accessToken, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate access token", "user_id", user.ID, "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Google login successful", "user_id", user.ID, "is_new_user", isNewUser)

	return &domain.AuthResponse{
		AccessToken: accessToken,
		IsNewUser:   isNewUser,
	}, nil
}

func (u *authUsecase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, userID)
}
