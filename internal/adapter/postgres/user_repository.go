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

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

type userRepository struct {
	db postgres.Client
}

func NewUserRepository(db postgres.Client) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
		SELECT id, email, password_hash, username, birth_date, is_confirmed, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.db.QueryRow(ctx, q, email)

	u := &models.User{}
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Username,
		&u.BirthDate, &u.IsConfirmed, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	const q = `
		INSERT INTO users (id, email, password_hash, username, birth_date, is_confirmed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx, q,
		user.ID, user.Email, user.PasswordHash, user.Username,
		user.BirthDate, user.IsConfirmed, user.CreatedAt, user.UpdatedAt,
	)
	return err
}
