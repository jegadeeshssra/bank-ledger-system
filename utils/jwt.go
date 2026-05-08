package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// In production, this should fail or use a secure default
		// For development, we'll use a default but log a warning
		fmt.Println("WARNING: JWT_SECRET environment variable not set. Using default secret for development.")
		secret = "your-development-secret-key-change-in-production"
	}
	jwtSecret = []byte(secret)
}

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token with user ID as subject
func GenerateJWT(userID uuid.UUID) (string, error) {
	now := time.Now()
	expirationTime := now.Add(24 * time.Hour) // Token expires in 24 hours

	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(), // Use user ID as subject
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ledger-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates the JWT token and returns the user ID
func ValidateJWT(tokenString string) (uuid.UUID, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid JWT token")
	}

	// Additional validation
	if claims.Subject == "" {
		return uuid.Nil, fmt.Errorf("missing subject in JWT token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in JWT token: %w", err)
	}

	return userID, nil
}
