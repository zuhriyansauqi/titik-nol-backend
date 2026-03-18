package usecase

import (
	"context"

	"github.com/mzhryns/titik-nol-backend/internal/domain"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) Create(ctx context.Context, user *domain.User) error {
	// Business logic: check if email already exists
	existingUser, _ := u.userRepo.GetByEmail(ctx, user.Email)
	if existingUser != nil {
		return domain.ErrEmailAlreadyExists
	}

	return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUsecase) Fetch(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.Fetch(ctx)
}
