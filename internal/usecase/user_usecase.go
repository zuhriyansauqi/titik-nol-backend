package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) Create(ctx context.Context, user *domain.User) error {
	slog.InfoContext(ctx, "Creating user", "email", user.Email)

	// Business logic: check if email already exists
	existingUser, _ := u.userRepo.GetByEmail(ctx, user.Email)
	if existingUser != nil {
		slog.WarnContext(ctx, "Email already exists", "email", user.Email)
		return domain.ErrEmailAlreadyExists
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "email", user.Email, "error", err)
		return err
	}

	slog.InfoContext(ctx, "User created successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

func (u *userUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	slog.InfoContext(ctx, "Fetching user by ID", "user_id", id)

	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		slog.DebugContext(ctx, "User not found", "user_id", id)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) Fetch(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedResult, error) {
	users, total, err := u.userRepo.Fetch(ctx, params)
	if err != nil {
		return nil, err
	}

	totalPages := total / params.PerPage
	if total%params.PerPage != 0 {
		totalPages++
	}

	return &domain.PaginatedResult{
		Items:      users,
		TotalItems: total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}
