package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID     `json:"id" validate:"required"`
	UserID    uuid.NullUUID `json:"user_id"`
	Name      string        `json:"name" validate:"required"`
	Balance   string        `json:"balance" validate:"required,numeric"`
	Currency  string        `json:"currency" validate:"required"`
	IsSystem  bool          `json:"is_system"`
	CreatedAt time.Time     `json:"created_at"`
}
