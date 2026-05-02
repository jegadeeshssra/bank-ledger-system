package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID     `json:"id"`
	OwnerID   uuid.NullUUID `json:"owner_id"`
	Name      string        `json:"name"`
	Balance   string        `json:"balance"`
	Currency  string        `json:"currency"`
	IsSystem  bool          `json:"is_system"`
	CreatedAt sql.NullTime  `json:"created_at"`
}

type AccountRepository struct {
	DB *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

func (r *AccountRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY,
		owner_id UUID,
		name TEXT NOT NULL,
		balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		currency TEXT NOT NULL,
		is_system BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMP DEFAULT now()
	);`
	_, err := r.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating accounts table: %w", err)
	}
	return nil
}

func (r *AccountRepository) CreateAccount(acc *Account) error {
	query := `INSERT INTO accounts (id, owner_id, name, balance, currency, is_system, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.Exec(query, acc.ID, acc.OwnerID, acc.Name, acc.Balance, acc.Currency, acc.IsSystem, acc.CreatedAt)
	if err != nil {
		return fmt.Errorf("error inserting account: %w", err)
	}
	return nil
}

func (r *AccountRepository) GetAccount(id uuid.UUID) (*Account, error) {
	query := `SELECT id, owner_id, name, balance, currency, is_system, created_at FROM accounts WHERE id = $1`
	var acc Account
	err := r.DB.QueryRow(query, id).Scan(
		&acc.ID, &acc.OwnerID, &acc.Name, &acc.Balance, &acc.Currency, &acc.IsSystem, &acc.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if no record is found
		}
		return nil, fmt.Errorf("error fetching account: %w", err)
	}
	return &acc, nil
}

func (r *AccountRepository) ListAccounts() ([]Account, error) {
	query := `SELECT id, owner_id, name, balance, currency, is_system, created_at FROM accounts`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.ID, &acc.OwnerID, &acc.Name, &acc.Balance, &acc.Currency, &acc.IsSystem, &acc.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning account: %w", err)
		}
		accounts = append(accounts, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}
	return accounts, nil
}

func (r *AccountRepository) DeleteAccount(id uuid.UUID) error {
	query := `DELETE FROM accounts WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting account: %w", err)
	}
	return nil
}

func (r *AccountRepository) UpdateBalance(tx *sql.Tx, accountID uuid.UUID, amount string, increment bool) error {
	var query string
	if increment {
		query = `UPDATE accounts SET balance = balance + $1 WHERE id = $2`
	} else {
		query = `UPDATE accounts SET balance = balance - $1 WHERE id = $2`
	}

	var err error
	if tx != nil {
		_, err = tx.Exec(query, amount, accountID)
	}
	if err != nil {
		return fmt.Errorf("error updating balance: %w", err)
	}
	return nil
}

func (r *AccountRepository) CalculateBalance(accountID uuid.UUID) (string, error) {
	// Let's compute credits - debits
	query := `SELECT COALESCE(SUM(credit - debit), 0) FROM entries WHERE account_id = $1`
	var balance string
	err := r.DB.QueryRow(query, accountID).Scan(&balance)
	if err != nil {
		return "0.00", err
	}
	return balance, nil
}

// SetBalance overwrites the balance with exact amount
func (r *AccountRepository) SetBalance(accountID uuid.UUID, balance string) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	_, err := r.DB.Exec(query, balance, accountID)
	return err
}
