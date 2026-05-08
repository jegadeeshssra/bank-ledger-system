package repository

import (
	"database/sql"
	"fmt"

	"ledger-system/models"
	"ledger-system/utils"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		salt TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`
	_, err := r.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}
	return nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	salt := utils.GenerateSalt()
	passwordHash := utils.HashPasswordWithSalt(user.Password, salt)
	user.Salt = salt
	query := `INSERT INTO users (id, username, password_hash, salt, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(query, user.ID, user.Username, passwordHash, salt, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, salt, created_at FROM users WHERE username = $1`
	var user models.User
	var passwordHash string
	err := r.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &passwordHash, &user.Salt, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}
	user.Password = passwordHash // Store hash in Password field for verification
	return &user, nil
}

func (r *UserRepository) VerifyPassword(username, password string) (*models.User, error) {
	user, err := r.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	if !utils.CheckPasswordWithSalt(password, user.Password, user.Salt) {
		return nil, nil
	}
	user.Password = "xxxxxxxxxxxx" // Clear password hash
	return user, nil
}
