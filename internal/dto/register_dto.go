package dto

import "time"

type RegisterInput struct {
	Email     string
	Password  string
	Username  string
	BirthDate time.Time
}
