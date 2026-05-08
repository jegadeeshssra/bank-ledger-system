package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 32
)

// GenerateSalt generates a random salt for a user.
func GenerateSalt() string {
	salt := make([]byte, saltLen)
	_, err := rand.Read(salt)
	if err != nil {
		panic("failed to generate salt")
	}
	return hex.EncodeToString(salt)
}

// HashPasswordWithSalt hashes the password using Argon2id and the provided salt.
func HashPasswordWithSalt(password, saltHex string) string {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		panic("invalid salt format")
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("%x", hash)
}

// CheckPasswordWithSalt verifies a password against the stored hash and salt.
func CheckPasswordWithSalt(password, hash, saltHex string) bool {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}
	expectedHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("%x", expectedHash) == hash
}
