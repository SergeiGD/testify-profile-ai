package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/SergeiGD/testify-profile/internal/models"
	"github.com/SergeiGD/testify-profile/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ConfirmationRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.RegistrationConfirmation, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*models.RegistrationConfirmation, error)
	Create(ctx context.Context, confirmation *models.RegistrationConfirmation) error
	Update(ctx context.Context, confirmation *models.RegistrationConfirmation) error
	ConfirmUser(ctx context.Context, userID uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type confirmationRepository struct {
	db postgres.Client
}

func NewConfirmationRepository(db postgres.Client) ConfirmationRepository {
	return &confirmationRepository{db: db}
}

func (r *confirmationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.RegistrationConfirmation, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, last_sent_at, created_at, updated_at
		FROM registration_confirmations
		WHERE user_id = $1
	`
	row := r.db.QueryRow(ctx, q, userID)
	return scanConfirmation(row)
}

func (r *confirmationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.RegistrationConfirmation, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, last_sent_at, created_at, updated_at
		FROM registration_confirmations
		WHERE token_hash = $1
	`
	row := r.db.QueryRow(ctx, q, tokenHash)
	return scanConfirmation(row)
}

func (r *confirmationRepository) Create(ctx context.Context, confirmation *models.RegistrationConfirmation) error {
	const q = `
		INSERT INTO registration_confirmations
			(user_id, token_hash, expires_at, last_sent_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	now := time.Now()
	confirmation.CreatedAt = now
	confirmation.UpdatedAt = now

	_, err := r.db.Exec(ctx, q,
		confirmation.UserID, confirmation.TokenHash,
		confirmation.ExpiresAt, confirmation.LastSentAt,
		confirmation.CreatedAt, confirmation.UpdatedAt,
	)
	return err
}

func (r *confirmationRepository) Update(ctx context.Context, confirmation *models.RegistrationConfirmation) error {
	const q = `
		UPDATE registration_confirmations
		SET token_hash = $1, expires_at = $2, last_sent_at = $3, updated_at = $4
		WHERE id = $5
	`
	confirmation.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, q,
		confirmation.TokenHash, confirmation.ExpiresAt,
		confirmation.LastSentAt, confirmation.UpdatedAt,
		confirmation.ID,
	)
	return err
}

func (r *confirmationRepository) ConfirmUser(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE users SET is_confirmed = TRUE, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

func (r *confirmationRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM registration_confirmations WHERE user_id = $1`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

func scanConfirmation(row pgx.Row) (*models.RegistrationConfirmation, error) {
	c := &models.RegistrationConfirmation{}
	err := row.Scan(
		&c.ID, &c.UserID, &c.TokenHash,
		&c.ExpiresAt, &c.LastSentAt,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConfirmationNotFound
		}
		return nil, err
	}
	return c, nil
}
