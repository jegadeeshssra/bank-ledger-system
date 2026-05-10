package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ledger-system/middleware"
	"ledger-system/models"

	"github.com/google/uuid"
)

func (s *Server) GetAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid account id"}`, http.StatusBadRequest)
		return
	}

	// Get account directly by user ID and account ID (ensures ownership)
	acc, err := s.accRepo.GetAccountByUserID(accountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, `{"error": "Account not found"}`, http.StatusNotFound)
		return
	}

	if err := ValidateResponse(w, acc); err != nil {
		return
	}
	json.NewEncoder(w).Encode(acc)
}

func (s *Server) ListAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	accounts, err := s.accRepo.ListAccountsByUserID(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if accounts == nil {
		accounts = []models.Account{}
	}
	if err := ValidateResponse(w, accounts); err != nil {
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

func (s *Server) GetEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid account id"}`, http.StatusBadRequest)
		return
	}

	// Verify account ownership by attempting to get it
	acc, err := s.accRepo.GetAccountByUserID(accountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, `{"error": "Account not found or access denied"}`, http.StatusNotFound)
		return
	}

	entries, err := s.entryRepo.GetEntriesByAccountID(accountID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []models.Entry{}
	}
	if err := ValidateResponse(w, entries); err != nil {
		return
	}
	json.NewEncoder(w).Encode(entries)
}

// The term reconcile means to make two things consistent that should be equal but might have drifted apart.
func (s *Server) Reconcile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid account id"}`, http.StatusBadRequest)
		return
	}

	// Verify account ownership by attempting to get it
	acc, err := s.accRepo.GetAccountByUserID(accountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, `{"error": "Account not found or access denied"}`, http.StatusNotFound)
		return
	}

	balance, err := s.accRepo.CalculateBalance(accountID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	err = s.accRepo.SetBalance(accountID, balance)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Reconciled", "balance": balance})
}

func (s *Server) GetTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid account id"}`, http.StatusBadRequest)
		return
	}

	// Verify account ownership by attempting to get it
	acc, err := s.accRepo.GetAccountByUserID(accountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, `{"error": "Account not found or access denied"}`, http.StatusNotFound)
		return
	}

	trxStr := r.PathValue("transaction_id")
	trxID, err := uuid.Parse(trxStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid transaction id"}`, http.StatusBadRequest)
		return
	}

	entries, err := s.entryRepo.GetEntriesByTransactionAndAccountID(trxID, accountID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(entries) == 0 {
		http.Error(w, `{"error": "Transaction not found or does not belong to this account"}`, http.StatusNotFound)
		return
	}
	if err := ValidateResponse(w, entries); err != nil {
		return
	}

	json.NewEncoder(w).Encode(entries)
}
