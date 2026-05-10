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
	userRepo *repository.UserRepository
}

func NewAuthServer(userRepo *repository.UserRepository) *AuthServer {
	return &AuthServer{userRepo: userRepo}
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

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Error generating authentication token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token": token,
		"user_id":      user.ID,
		"username":     user.Username,
		"message":      "Login successful",
		"expires_in":   "24h",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
