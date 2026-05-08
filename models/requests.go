package models

import (
	"github.com/google/uuid"
)

type CreateAccountReq struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Currency string `json:"currency" validate:"required,len=3"`
	IsSystem bool   `json:"is_system"`
}

type AmountReq struct {
	Amount string `json:"amount" validate:"required,numeric,gt=0"`
}

type TransferReq struct {
	FromAccountID uuid.UUID `json:"from_account_id" validate:"required"`
	ToAccountID   uuid.UUID `json:"to_account_id" validate:"required"`
	Amount        string    `json:"amount" validate:"required,numeric,gt=0"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=12"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
