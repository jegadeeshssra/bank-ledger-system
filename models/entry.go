package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID            uuid.UUID      `json:"id" validate:"required"`
	AccountID     uuid.UUID      `json:"account_id" validate:"required"`
	UserID        uuid.UUID      `json:"user_id" validate:"required"`
	Debit         string         `json:"debit" validate:"required,numeric"`
	Credit        string         `json:"credit" validate:"required,numeric"`
	TransactionID uuid.UUID      `json:"transaction_id" validate:"required"`
	OperationType string         `json:"operation_type" validate:"required"`
	Description   sql.NullString `json:"description"`
	CreatedAt     time.Time      `json:"created_at"`
}
