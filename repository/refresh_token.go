package repository

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"ledger-system/models"

	"github.com/google/uuid"
)

type RefreshTokenRepository struct {
	DB *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{DB: db}
}

func (r *RefreshTokenRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		userid UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT NOT NULL,
		issued_at TIMESTAMP NOT NULL DEFAULT NOW(),
		expires_at TIMESTAMP NOT NULL,
		revoked BOOLEAN NOT NULL DEFAULT FALSE
	);`
	_, err := r.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating refresh_tokens table: %w", err)
	}
	return nil
}

// HashToken generates SHA256 hash of the token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// CreateRefreshToken stores a hashed refresh token in the database
func (r *RefreshTokenRepository) CreateRefreshToken(userID uuid.UUID, tokenString string, expiresAt time.Time) (*models.RefreshToken, error) {
	tokenHash := HashToken(tokenString)
	id := uuid.New()
	issuedAt := time.Now()

	query := `INSERT INTO refresh_tokens (id, userid, token_hash, issued_at, expires_at, revoked) 
	        VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Exec(query, id, userID, tokenHash, issuedAt, expiresAt, false)
	if err != nil {
		return nil, fmt.Errorf("error creating refresh token: %w", err)
	}

	return &models.RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}, nil
}

// GetValidRefreshToken retrieves a refresh token if it's valid (not expired and not revoked)
func (r *RefreshTokenRepository) GetValidRefreshToken(userID uuid.UUID, tokenString string) (*models.RefreshToken, error) {
	tokenHash := HashToken(tokenString)

	query := `SELECT id, userid, token_hash, issued_at, expires_at, revoked 
	        FROM refresh_tokens 
	        WHERE userid = $1 AND token_hash = $2 AND expires_at > NOW() AND revoked = FALSE`

	var rt models.RefreshToken
	err := r.DB.QueryRow(query, userID, tokenHash).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.IssuedAt,
		&rt.ExpiresAt,
		&rt.Revoked,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching refresh token: %w", err)
	}

	return &rt, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *RefreshTokenRepository) RevokeRefreshToken(tokenID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`
	_, err := r.DB.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("error revoking refresh token: %w", err)
	}
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllUserTokens(userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE userid = $1 AND revoked = FALSE`
	_, err := r.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error revoking user tokens: %w", err)
	}
	return nil
}
