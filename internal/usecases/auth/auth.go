package auth

import (
	"context"
	"errors"

	"github.com/SergeiGD/testify-profile/internal/adapter/postgres"
	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/SergeiGD/testify-profile/internal/dto"
	jwtSvc "github.com/SergeiGD/testify-profile/internal/services/jwt"
	"github.com/SergeiGD/testify-profile/internal/services/password"
	"github.com/google/uuid"
)

type AuthUseCase interface {
	Login(ctx context.Context, input dto.LoginInput) (*dto.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
}

type authUseCase struct {
	userRepo postgres.UserRepository
	hasher   password.PasswordHasher
	jwtSvc   jwtSvc.JWTService
}

func NewAuthUseCase(
	userRepo postgres.UserRepository,
	hasher password.PasswordHasher,
	jwtSvc jwtSvc.JWTService,
) AuthUseCase {
	return &authUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		jwtSvc:   jwtSvc,
	}
}

func (uc *authUseCase) Login(ctx context.Context, input dto.LoginInput) (*dto.TokenPair, error) {
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsConfirmed {
		return nil, domain.ErrAccountNotConfirmed
	}

	if err := uc.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return uc.issueTokenPair(user.ID, user.Username, user.BirthDate.Format("2006-01-02"))
}

func (uc *authUseCase) Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error) {
	claims, err := uc.jwtSvc.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidRefreshToken
		}
		return nil, err
	}

	if !user.IsConfirmed {
		return nil, domain.ErrAccountNotConfirmed
	}

	return uc.issueTokenPair(user.ID, user.Username, user.BirthDate.Format("2006-01-02"))
}

func (uc *authUseCase) issueTokenPair(userID uuid.UUID, username, birthDate string) (*dto.TokenPair, error) {
	accessToken, err := uc.jwtSvc.GenerateAccessToken(userID, username, birthDate)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtSvc.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
