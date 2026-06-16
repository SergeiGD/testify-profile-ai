package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Username     string
	BirthDate    time.Time
	IsConfirmed  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
