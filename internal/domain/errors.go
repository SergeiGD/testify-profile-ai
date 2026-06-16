package domain

import "errors"

var (
	ErrEmailAlreadyConfirmed    = errors.New("email already confirmed")
	ErrResendTooFrequent        = errors.New("confirmation email was sent recently, please wait before requesting again")
	ErrInvalidOrExpiredToken    = errors.New("invalid or expired confirmation token")
	ErrUserNotFound             = errors.New("user not found")
	ErrConfirmationNotFound     = errors.New("confirmation not found")
)
