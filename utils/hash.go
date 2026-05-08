package utils

import (
	"fmt"

	"golang.org/x/crypto/argon2"
)

// HashPassword hashes a password using Argon2id
func HashPassword(password string) string {
	salt := []byte("some-random-salt") // In production, use a unique salt per user
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x", hash)
}

// CheckPassword verifies a password against a hash
func CheckPassword(password, hash string) bool {
	salt := []byte("some-random-salt")
	expectedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x", expectedHash) == hash
}
