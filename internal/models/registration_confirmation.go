package models

import (
	"time"

	"github.com/google/uuid"
)

type RegistrationConfirmation struct {
	ID         int
	UserID     uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	LastSentAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
