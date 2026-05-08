package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ledger-system/models"
	"ledger-system/repository"

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

	// For simplicity, return user ID. In production, use JWT or session.
	response := map[string]interface{}{
		"user_id": user.ID,
		"message": "Login successful",
	}
	json.NewEncoder(w).Encode(response)
}
