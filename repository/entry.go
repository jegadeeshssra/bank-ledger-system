package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID            uuid.UUID      `json:"id" validate:"required"`
	AccountID     uuid.UUID      `json:"account_id" validate:"required"`
	Debit         string         `json:"debit" validate:"required,numeric"`
	Credit        string         `json:"credit" validate:"required,numeric"`
	TransactionID uuid.UUID      `json:"transaction_id" validate:"required"`
	OperationType string         `json:"operation_type" validate:"required"`
	Description   sql.NullString `json:"description"`
	CreatedAt     time.Time      `json:"created_at"`
}

type EntryRepository struct {
	DB *sql.DB
}

func NewEntryRepository(db *sql.DB) *EntryRepository {
	return &EntryRepository{DB: db}
}

func (r *EntryRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS entries (
		id UUID PRIMARY KEY,
		account_id UUID NOT NULL,
		debit DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		credit DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		transaction_id UUID NOT NULL,
		operation_type TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT now(),
		CONSTRAINT fk_account
			FOREIGN KEY(account_id) 
			REFERENCES accounts(id)
	);`
	_, err := r.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating entries table: %w", err)
	}
	return nil
}

func (r *EntryRepository) InsertEntry(tx *sql.Tx, entry *Entry) error {
	query := `INSERT INTO entries (id, account_id, debit, credit, transaction_id, operation_type, description, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	var err error
	if tx != nil {
		_, err = tx.Exec(query, entry.ID, entry.AccountID, entry.Debit, entry.Credit, entry.TransactionID, entry.OperationType, entry.Description, entry.CreatedAt)
	}
	if err != nil {
		return fmt.Errorf("error inserting entry: %w", err)
	}
	return nil
}

func (r *EntryRepository) GetEntriesByAccountID(accountID uuid.UUID) ([]Entry, error) {
	query := `SELECT id, account_id, debit, credit, transaction_id, operation_type, description, created_at 
	          FROM entries WHERE account_id = $1 ORDER BY created_at ASC`
	rows, err := r.DB.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("error querying entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(
			&entry.ID, &entry.AccountID, &entry.Debit, &entry.Credit,
			&entry.TransactionID, &entry.OperationType, &entry.Description, &entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}
	return entries, nil
}

func (r *EntryRepository) GetEntriesByTransactionID(transactionID uuid.UUID) ([]Entry, error) {
	query := `SELECT id, account_id, debit, credit, transaction_id, operation_type, description, created_at 
	          FROM entries WHERE transaction_id = $1`
	rows, err := r.DB.Query(query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("error querying entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(
			&entry.ID, &entry.AccountID, &entry.Debit, &entry.Credit,
			&entry.TransactionID, &entry.OperationType, &entry.Description, &entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *EntryRepository) GetEntriesByTransactionAndAccountID(transactionID, accountID uuid.UUID) ([]Entry, error) {
	query := `SELECT id, account_id, debit, credit, transaction_id, operation_type, description, created_at 
	          FROM entries WHERE transaction_id = $1 AND account_id = $2`
	rows, err := r.DB.Query(query, transactionID, accountID)
	if err != nil {
		return nil, fmt.Errorf("error querying entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(
			&entry.ID, &entry.AccountID, &entry.Debit, &entry.Credit,
			&entry.TransactionID, &entry.OperationType, &entry.Description, &entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
