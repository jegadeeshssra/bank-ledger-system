package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	Username  string    `json:"username" validate:"required,min=3,max=50"`
	Password  string    `json:"-"` // Never serialize password
	Salt      string    `json:"-"` // Salt for password hashing
	CreatedAt time.Time `json:"created_at"`
}
