package profile

import (
	"context"
	"errors"

	"github.com/SergeiGD/testify-profile/internal/adapter/postgres"
	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/SergeiGD/testify-profile/internal/models"
	"github.com/google/uuid"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type profileUseCase struct {
	userRepo postgres.UserRepository
}

func NewProfileUseCase(userRepo postgres.UserRepository) ProfileUseCase {
	return &profileUseCase{userRepo: userRepo}
}

func (uc *profileUseCase) GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	if !user.IsConfirmed {
		return nil, domain.ErrAccountNotConfirmed
	}

	return user, nil
}
