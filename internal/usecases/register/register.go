package register

import (
	"context"
	"errors"
	"time"

	pgadapter "github.com/SergeiGD/testify-profile/internal/adapter/postgres"
	smtpadapter "github.com/SergeiGD/testify-profile/internal/adapter/smtp"
	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/SergeiGD/testify-profile/internal/dto"
	"github.com/SergeiGD/testify-profile/internal/models"
	"github.com/SergeiGD/testify-profile/internal/services/clock"
	"github.com/SergeiGD/testify-profile/internal/services/linkbuilder"
	"github.com/SergeiGD/testify-profile/internal/services/password"
	"github.com/SergeiGD/testify-profile/internal/services/token"
)

type RegisterUseCase interface {
	Register(ctx context.Context, input dto.RegisterInput) error
	Confirm(ctx context.Context, rawToken string) error
}

type registerUseCase struct {
	userRepo         pgadapter.UserRepository
	confirmationRepo pgadapter.ConfirmationRepository
	emailSender      smtpadapter.EmailSender
	passwordHasher   password.PasswordHasher
	tokenService     token.TokenService
	linkBuilder      linkbuilder.LinkBuilder
	clock            clock.Clock
	tokenTTL         time.Duration
	resendCooldown   time.Duration
}

func NewRegisterUseCase(
	userRepo pgadapter.UserRepository,
	confirmationRepo pgadapter.ConfirmationRepository,
	emailSender smtpadapter.EmailSender,
	passwordHasher password.PasswordHasher,
	tokenService token.TokenService,
	linkBuilder linkbuilder.LinkBuilder,
	clock clock.Clock,
	tokenTTL time.Duration,
	resendCooldown time.Duration,
) RegisterUseCase {
	return &registerUseCase{
		userRepo:         userRepo,
		confirmationRepo: confirmationRepo,
		emailSender:      emailSender,
		passwordHasher:   passwordHasher,
		tokenService:     tokenService,
		linkBuilder:      linkBuilder,
		clock:            clock,
		tokenTTL:         tokenTTL,
		resendCooldown:   resendCooldown,
	}
}

func (uc *registerUseCase) Register(ctx context.Context, input dto.RegisterInput) error {
	existingUser, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}

	if existingUser != nil {
		if existingUser.IsConfirmed {
			return domain.ErrEmailAlreadyConfirmed
		}
		return uc.handleResend(ctx, existingUser)
	}

	return uc.handleNewUser(ctx, input)
}

func (uc *registerUseCase) handleNewUser(ctx context.Context, input dto.RegisterInput) error {
	passwordHash, err := uc.passwordHasher.Hash(input.Password)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: passwordHash,
		Username:     input.Username,
		BirthDate:    input.BirthDate,
		IsConfirmed:  false,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return err
	}

	return uc.createAndSendConfirmation(ctx, user)
}

func (uc *registerUseCase) handleResend(ctx context.Context, user *models.User) error {
	confirmation, err := uc.confirmationRepo.FindByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, domain.ErrConfirmationNotFound) {
		return err
	}

	now := uc.clock.Now()

	if confirmation != nil {
		if now.Before(confirmation.LastSentAt.Add(uc.resendCooldown)) {
			return domain.ErrResendTooFrequent
		}

		rawToken, tokenHash, err := uc.tokenService.Generate()
		if err != nil {
			return err
		}

		confirmation.TokenHash = tokenHash
		confirmation.ExpiresAt = now.Add(uc.tokenTTL)
		confirmation.LastSentAt = now

		if err := uc.confirmationRepo.Update(ctx, confirmation); err != nil {
			return err
		}

		link := uc.linkBuilder.ConfirmationLink(rawToken)
		return uc.emailSender.SendConfirmationEmail(user.Email, link)
	}

	return uc.createAndSendConfirmation(ctx, user)
}

func (uc *registerUseCase) createAndSendConfirmation(ctx context.Context, user *models.User) error {
	rawToken, tokenHash, err := uc.tokenService.Generate()
	if err != nil {
		return err
	}

	now := uc.clock.Now()
	confirmation := &models.RegistrationConfirmation{
		UserID:     user.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  now.Add(uc.tokenTTL),
		LastSentAt: now,
	}

	if err := uc.confirmationRepo.Create(ctx, confirmation); err != nil {
		return err
	}

	link := uc.linkBuilder.ConfirmationLink(rawToken)
	return uc.emailSender.SendConfirmationEmail(user.Email, link)
}

func (uc *registerUseCase) Confirm(ctx context.Context, rawToken string) error {
	tokenHash := uc.tokenService.Hash(rawToken)

	confirmation, err := uc.confirmationRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrConfirmationNotFound) {
			return domain.ErrInvalidOrExpiredToken
		}
		return err
	}

	if uc.clock.Now().After(confirmation.ExpiresAt) {
		return domain.ErrInvalidOrExpiredToken
	}

	if err := uc.confirmationRepo.ConfirmUser(ctx, confirmation.UserID); err != nil {
		return err
	}

	return uc.confirmationRepo.DeleteByUserID(ctx, confirmation.UserID)
}
