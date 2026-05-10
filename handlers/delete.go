package handlers

import (
	"fmt"
	"net/http"

	"ledger-system/middleware"

	"github.com/google/uuid"
)

func (s *Server) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid account id"}`, http.StatusBadRequest)
		return
	}

	// Check if the account belongs to the authenticated user
	acc, err := s.accRepo.GetAccountByUserID(id, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, `{"error": "Account not found or access denied"}`, http.StatusNotFound)
		return
	}

	err = s.accRepo.DeleteAccount(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
