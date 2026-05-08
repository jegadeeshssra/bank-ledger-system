package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"ledger-system/utils"

	"github.com/google/uuid"
)

// JWTMiddleware validates JWT tokens and adds user ID to request context
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Invalid authorization header format. Use 'Bearer <token>'"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http.Error(w, `{"error": "Token is required"}`, http.StatusUnauthorized)
			return
		}

		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Invalid token: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	return userID, ok
}
