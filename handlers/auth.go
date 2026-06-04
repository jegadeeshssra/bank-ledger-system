package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ledger-system/models"
	"ledger-system/repository"
	"ledger-system/utils"

	"github.com/google/uuid"
)

type AuthServer struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
}

func NewAuthServer(userRepo *repository.UserRepository, refreshTokenRepo *repository.RefreshTokenRepository) *AuthServer {
	return &AuthServer{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (a *AuthServer) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(w, req); err != nil {
		return
	}

	// Check if user already exists
	existingUser, err := a.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	user := &models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Password:  req.Password,
		CreatedAt: time.Now(),
	}

	err = a.userRepo.CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.Password = "" // Don't return password
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (a *AuthServer) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(w, req); err != nil {
		return
	}

	user, err := a.userRepo.VerifyPassword(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate access token
	accessToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshTokenString, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	// Store refresh token in database
	refreshTokenExpiry := time.Now().Add(utils.GetRefreshTokenExpiry())
	_, err = a.refreshTokenRepo.CreateRefreshToken(user.ID, refreshTokenString, refreshTokenExpiry)
	if err != nil {
		http.Error(w, "Error storing refresh token", http.StatusInternalServerError)
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    "5m",
		TokenType:    "Bearer",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Refresh generates a new access token using a valid refresh token
func (a *AuthServer) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// validate the request with the struct tags
	if err := ValidateRequest(w, req); err != nil {
		return
	}

	// Extract user ID from access token in Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error": "Missing authorization header"}`, http.StatusUnauthorized)
		return
	}

	// Parse bearer token
	var tokenString string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		http.Error(w, `{"error": "Invalid authorization header format"}`, http.StatusUnauthorized)
		return
	}

	// Get user ID from token (even if expired, we still want the user ID)
	userID, err := utils.ValidateJWT(tokenString)
	if err != nil {
		// Token might be expired, but we still need the user ID from claims
		// Let's use a function that parses without validating expiry
		userID, err = utils.ExtractUserIDFromToken(tokenString)
		if err != nil {
			http.Error(w, `{"error": "Invalid access token"}`, http.StatusUnauthorized)
			return
		}
	}

	// Verify the refresh token is valid
	refreshToken, err := a.refreshTokenRepo.GetValidRefreshToken(userID, req.RefreshToken)
	if err != nil {
		// Revoke all tokens for this user due to invalid token
		a.refreshTokenRepo.RevokeAllUserTokens(userID)
		http.Error(w, `{"error": "Invalid refresh token"}`, http.StatusInternalServerError)
		return
	}
	if refreshToken == nil {
		// Revoke all tokens for this user due to expired or not found token
		a.refreshTokenRepo.RevokeAllUserTokens(userID)
		http.Error(w, `{"error": "Refresh token not found or expired"}`, http.StatusUnauthorized)
		return
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateJWT(userID)
	if err != nil {
		http.Error(w, `{"error": "Error generating new access token"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token": newAccessToken,
		"expires_in":   "5m",
		"token_type":   "Bearer",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
