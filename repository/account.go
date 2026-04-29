package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type Account struct {
	ID        int       `json:"id"`
	Owner     string    `json:"owner"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
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
		id SERIAL PRIMARY KEY,
		owner TEXT NOT NULL,
		balance BIGINT NOT NULL,
		currency TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT now()
	);`
	_, err := r.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}
	return nil
}

func (r *AccountRepository) CreateAccount(owner string, balance int64, currency string) (*Account, error) {
	query := `INSERT INTO accounts (owner, balance, currency) VALUES ($1, $2, $3) RETURNING id, owner, balance, currency, created_at`
	var acc Account
	err := r.DB.QueryRow(query, owner, balance, currency).Scan(
		&acc.ID, &acc.Owner, &acc.Balance, &acc.Currency, &acc.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error inserting account: %w", err)
	}
	return &acc, nil
}

func (r *AccountRepository) GetAccount(id int) (*Account, error) {
	query := `SELECT id, owner, balance, currency, created_at FROM accounts WHERE id = $1`
	var acc Account
	err := r.DB.QueryRow(query, id).Scan(
		&acc.ID, &acc.Owner, &acc.Balance, &acc.Currency, &acc.CreatedAt,
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
	query := `SELECT id, owner, balance, currency, created_at FROM accounts`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.ID, &acc.Owner, &acc.Balance, &acc.Currency, &acc.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning account: %w", err)
		}
		accounts = append(accounts, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}
	return accounts, nil
}

func (r *AccountRepository) DeleteAccount(id int) error {
	query := `DELETE FROM accounts WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting account: %w", err)
	}
	return nil
}
