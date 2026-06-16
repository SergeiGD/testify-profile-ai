package register_test

import (
	"context"
	"testing"
	"time"

	pgmocks "github.com/SergeiGD/testify-profile/internal/adapter/postgres/mocks"
	smtpmocks "github.com/SergeiGD/testify-profile/internal/adapter/smtp/mocks"
	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/SergeiGD/testify-profile/internal/dto"
	"github.com/SergeiGD/testify-profile/internal/models"
	clockmocks "github.com/SergeiGD/testify-profile/internal/services/clock/mocks"
	"github.com/SergeiGD/testify-profile/internal/services/linkbuilder"
	"github.com/SergeiGD/testify-profile/internal/services/password"
	"github.com/SergeiGD/testify-profile/internal/services/token"
	"github.com/SergeiGD/testify-profile/internal/usecases/register"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildUseCase(
	t *testing.T,
	userRepo *pgmocks.MockUserRepository,
	confirmRepo *pgmocks.MockConfirmationRepository,
	emailSender *smtpmocks.MockEmailSender,
	clk *clockmocks.MockClock,
) register.RegisterUseCase {
	t.Helper()
	return register.NewRegisterUseCase(
		userRepo,
		confirmRepo,
		emailSender,
		password.NewPasswordHasher(),
		token.NewTokenService(),
		linkbuilder.NewLinkBuilder("http://localhost:8080"),
		clk,
		10*time.Minute,
		1*time.Minute,
	)
}

func defaultInput() dto.RegisterInput {
	return dto.RegisterInput{
		Email:     "test@example.com",
		Password:  "secret123",
		Username:  "testuser",
		BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestRegister(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		// setupMocks настраивает ожидания и возвращает необязательную функцию
		// для дополнительных проверок после вызова Register.
		setupMocks func(*pgmocks.MockUserRepository, *pgmocks.MockConfirmationRepository, *smtpmocks.MockEmailSender, *clockmocks.MockClock) func(*testing.T)
		wantErr    error
	}{
		{
			name: "new user: creates user and sends email",
			setupMocks: func(
				userRepo *pgmocks.MockUserRepository,
				confirmRepo *pgmocks.MockConfirmationRepository,
				emailSender *smtpmocks.MockEmailSender,
				clk *clockmocks.MockClock,
			) func(*testing.T) {
				var createdUser *models.User

				userRepo.EXPECT().
					FindByEmail(mock.Anything, "test@example.com").
					Return(nil, domain.ErrUserNotFound).
					Once()
				userRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*models.User")).
					Run(func(_ context.Context, u *models.User) { createdUser = u }).
					Return(nil).
					Once()
				clk.EXPECT().Now().Return(now).Once()
				confirmRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*models.RegistrationConfirmation")).
					Return(nil).
					Once()
				emailSender.EXPECT().
					SendConfirmationEmail("test@example.com", mock.AnythingOfType("string")).
					Return(nil).
					Once()

				return func(t *testing.T) {
					require.NotNil(t, createdUser)
					assert.False(t, createdUser.IsConfirmed, "new user should not be confirmed yet")
				}
			},
		},
		{
			name: "already confirmed email: returns error",
			setupMocks: func(
				userRepo *pgmocks.MockUserRepository,
				_ *pgmocks.MockConfirmationRepository,
				_ *smtpmocks.MockEmailSender,
				_ *clockmocks.MockClock,
			) func(*testing.T) {
				userRepo.EXPECT().
					FindByEmail(mock.Anything, "test@example.com").
					Return(&models.User{
						ID:          uuid.New(),
						Email:       "test@example.com",
						IsConfirmed: true,
					}, nil).
					Once()
				return nil
			},
			wantErr: domain.ErrEmailAlreadyConfirmed,
		},
		{
			name: "resend before cooldown: returns error",
			setupMocks: func(
				userRepo *pgmocks.MockUserRepository,
				confirmRepo *pgmocks.MockConfirmationRepository,
				_ *smtpmocks.MockEmailSender,
				clk *clockmocks.MockClock,
			) func(*testing.T) {
				userID := uuid.New()
				_, hash, _ := token.NewTokenService().Generate()

				userRepo.EXPECT().
					FindByEmail(mock.Anything, "test@example.com").
					Return(&models.User{ID: userID, Email: "test@example.com", IsConfirmed: false}, nil).
					Once()
				confirmRepo.EXPECT().
					FindByUserID(mock.Anything, userID).
				Return(&models.RegistrationConfirmation{
					ID:         1,
					UserID:     userID,
					TokenHash:  hash,
					ExpiresAt:  now.Add(10 * time.Minute),
					LastSentAt: now.Add(-30 * time.Second), // 30s ago — within 1-minute cooldown
				}, nil).
					Once()
				clk.EXPECT().Now().Return(now).Once()
				return nil
			},
			wantErr: domain.ErrResendTooFrequent,
		},
		{
			name: "resend after cooldown: sends email",
			setupMocks: func(
				userRepo *pgmocks.MockUserRepository,
				confirmRepo *pgmocks.MockConfirmationRepository,
				emailSender *smtpmocks.MockEmailSender,
				clk *clockmocks.MockClock,
			) func(*testing.T) {
				userID := uuid.New()
				_, hash, _ := token.NewTokenService().Generate()

				userRepo.EXPECT().
					FindByEmail(mock.Anything, "test@example.com").
					Return(&models.User{ID: userID, Email: "test@example.com", IsConfirmed: false}, nil).
					Once()
				confirmRepo.EXPECT().
					FindByUserID(mock.Anything, userID).
				Return(&models.RegistrationConfirmation{
					ID:         1,
					UserID:     userID,
					TokenHash:  hash,
					ExpiresAt:  now.Add(10 * time.Minute),
					LastSentAt: now.Add(-2 * time.Minute), // 2 minutes ago — past cooldown
				}, nil).
					Once()
				clk.EXPECT().Now().Return(now).Once()
				confirmRepo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*models.RegistrationConfirmation")).
					Return(nil).
					Once()
				emailSender.EXPECT().
					SendConfirmationEmail("test@example.com", mock.AnythingOfType("string")).
					Return(nil).
					Once()
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := pgmocks.NewMockUserRepository(t)
			confirmRepo := pgmocks.NewMockConfirmationRepository(t)
			emailSender := smtpmocks.NewMockEmailSender(t)
			clk := clockmocks.NewMockClock(t)

			afterAssert := tt.setupMocks(userRepo, confirmRepo, emailSender, clk)

			uc := buildUseCase(t, userRepo, confirmRepo, emailSender, clk)
			err := uc.Register(context.Background(), defaultInput())

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			if afterAssert != nil {
				afterAssert(t)
			}
		})
	}
}

func TestConfirm(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	tests := []struct {
		name string
		// setupMocks настраивает ожидания и возвращает токен для передачи в Confirm.
		setupMocks func(*pgmocks.MockConfirmationRepository, *clockmocks.MockClock) string
		wantErr    error
	}{
		{
			name: "valid token: confirms user",
			setupMocks: func(confirmRepo *pgmocks.MockConfirmationRepository, clk *clockmocks.MockClock) string {
				rawToken, hash, _ := token.NewTokenService().Generate()
				confirmRepo.EXPECT().
					FindByTokenHash(mock.Anything, hash).
				Return(&models.RegistrationConfirmation{
					ID:        1,
					UserID:    userID,
					TokenHash: hash,
					ExpiresAt: now.Add(10 * time.Minute),
				}, nil).
				Once()
			clk.EXPECT().Now().Return(now).Once()
			confirmRepo.EXPECT().
				ConfirmUser(mock.Anything, userID).
					Return(nil).
					Once()
				confirmRepo.EXPECT().
					DeleteByUserID(mock.Anything, userID).
					Return(nil).
					Once()
				return rawToken
			},
		},
		{
			name: "expired token: returns error",
			setupMocks: func(confirmRepo *pgmocks.MockConfirmationRepository, clk *clockmocks.MockClock) string {
				rawToken, hash, _ := token.NewTokenService().Generate()
				confirmRepo.EXPECT().
					FindByTokenHash(mock.Anything, hash).
				Return(&models.RegistrationConfirmation{
					ID:        1,
					UserID:    userID,
					TokenHash: hash,
					ExpiresAt: now.Add(-1 * time.Second), // already expired
				}, nil).
					Once()
				clk.EXPECT().Now().Return(now).Once()
				return rawToken
			},
			wantErr: domain.ErrInvalidOrExpiredToken,
		},
		{
			name: "invalid token: returns error",
			setupMocks: func(confirmRepo *pgmocks.MockConfirmationRepository, _ *clockmocks.MockClock) string {
				confirmRepo.EXPECT().
					FindByTokenHash(mock.Anything, mock.AnythingOfType("string")).
					Return(nil, domain.ErrConfirmationNotFound).
					Once()
				return "completely-invalid-token"
			},
			wantErr: domain.ErrInvalidOrExpiredToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := pgmocks.NewMockUserRepository(t)
			confirmRepo := pgmocks.NewMockConfirmationRepository(t)
			emailSender := smtpmocks.NewMockEmailSender(t)
			clk := clockmocks.NewMockClock(t)

			tok := tt.setupMocks(confirmRepo, clk)

			uc := buildUseCase(t, userRepo, confirmRepo, emailSender, clk)
			err := uc.Confirm(context.Background(), tok)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
