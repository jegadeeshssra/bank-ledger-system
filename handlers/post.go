package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ledger-system/middleware"
	"ledger-system/models"

	"github.com/google/uuid"
)

func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized - valid JWT token required"}`, http.StatusUnauthorized)
		return
	}

	var reqBody models.CreateAccountReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := ValidateRequest(w, reqBody); err != nil {
		return
	}

	acc := &models.Account{
		ID:        uuid.New(),
		UserID:    userID, // Assign authenticated user ID
		Name:      reqBody.Name,
		Balance:   "0.00",
		Currency:  reqBody.Currency,
		IsSystem:  reqBody.IsSystem,
		CreatedAt: time.Now(),
	}

	err := s.accRepo.CreateAccount(acc)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if err := ValidateResponse(w, acc); err != nil {
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acc)
}

func (s *Server) Deposit(w http.ResponseWriter, r *http.Request) {
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

	// Check if the account belongs to the authenticated user
	acc, err := s.accRepo.GetAccountByUserID(accountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, `{"error": "Account not found or access denied"}`, http.StatusNotFound)
		return
	}

	var reqBody models.AmountReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := ValidateRequest(w, reqBody); err != nil {
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, `{"error": "Failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	trxID := uuid.New()
	entry := &models.Entry{
		ID:            uuid.New(),
		AccountID:     accountID,
		UserID:        userID, // Set authenticated user ID
		Credit:        reqBody.Amount,
		Debit:         "0.00",
		TransactionID: trxID,
		OperationType: "DEPOSIT",
		CreatedAt:     time.Now(),
	}

	if err := s.entryRepo.InsertEntry(tx, entry); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if err := s.accRepo.UpdateBalance(tx, accountID, reqBody.Amount, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	if err := ValidateResponse(w, entry); err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entry)
}

func (s *Server) Withdraw(w http.ResponseWriter, r *http.Request) {
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

	// Check if the account belongs to the authenticated user
	acc, err := s.accRepo.GetAccountByUserID(accountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, `{"error": "Account not found or access denied"}`, http.StatusNotFound)
		return
	}

	var reqBody models.AmountReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := ValidateRequest(w, reqBody); err != nil {
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, `{"error": "Failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var sufficient bool
	err = tx.QueryRow("SELECT balance >= CAST($1 AS NUMERIC) FROM accounts WHERE id = $2 FOR UPDATE", reqBody.Amount, accountID).Scan(&sufficient)
	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve account or invalid amount"}`, http.StatusBadRequest)
		return
	}
	if !sufficient {
		http.Error(w, `{"error": "Insufficient funds"}`, http.StatusBadRequest)
		return
	}

	trxID := uuid.New()
	entry := &models.Entry{
		ID:            uuid.New(),
		AccountID:     accountID,
		UserID:        userID, // Set authenticated user ID
		Credit:        "0.00",
		Debit:         reqBody.Amount,
		TransactionID: trxID,
		OperationType: "WITHDRAW",
		CreatedAt:     time.Now(),
	}

	if err := s.entryRepo.InsertEntry(tx, entry); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if err := s.accRepo.UpdateBalance(tx, accountID, reqBody.Amount, false); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	// When you call tx.Commit() in a database transaction, the transaction is permanently saved to the database.
	// After that point, a Rollback() call does nothing because there's nothing left to undo.
	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error": "Failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	if err := ValidateResponse(w, entry); err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entry)
}

func (s *Server) Transfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	fromAccountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid account id"}`, http.StatusBadRequest)
		return
	}

	var reqBody models.TransferReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := ValidateRequest(w, reqBody); err != nil {
		return
	}

	if reqBody.FromAccountID != uuid.Nil && reqBody.FromAccountID != fromAccountID {
		http.Error(w, `{"error": "Mismatch between URL account ID and body from_account_id"}`, http.StatusBadRequest)
		return
	}
	if reqBody.FromAccountID == reqBody.ToAccountID {
		http.Error(w, `{"error": "from and to account cannot be same"}`, http.StatusBadRequest)
		return
	}

	reqBody.FromAccountID = fromAccountID

	// Check if both accounts belong to the authenticated user
	fromAcc, err := s.accRepo.GetAccountByUserID(reqBody.FromAccountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if fromAcc == nil {
		http.Error(w, `{"error": "Source account not found or access denied"}`, http.StatusNotFound)
		return
	}

	toAcc, err := s.accRepo.GetAccountByUserID(reqBody.ToAccountID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if toAcc == nil {
		http.Error(w, `{"error": "Destination account not found or access denied"}`, http.StatusNotFound)
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, `{"error": "Failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var sufficient bool
	err = tx.QueryRow("SELECT balance >= CAST($1 AS NUMERIC) FROM accounts WHERE id = $2 FOR UPDATE", reqBody.Amount, reqBody.FromAccountID).Scan(&sufficient)
	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve account or invalid amount"}`, http.StatusBadRequest)
		return
	}
	if !sufficient {
		http.Error(w, `{"error": "Insufficient funds"}`, http.StatusBadRequest)
		return
	}

	trxID := uuid.New()

	// From Account (Debit)
	entryFrom := &models.Entry{
		ID:            uuid.New(),
		AccountID:     reqBody.FromAccountID,
		UserID:        userID, // Set authenticated user ID
		Credit:        "0.00",
		Debit:         reqBody.Amount,
		TransactionID: trxID,
		OperationType: "TRANSFER_OUT",
		CreatedAt:     time.Now(),
	}
	if err := s.entryRepo.InsertEntry(tx, entryFrom); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if err := s.accRepo.UpdateBalance(tx, reqBody.FromAccountID, reqBody.Amount, false); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// To Account (Credit)
	entryTo := &models.Entry{
		ID:            uuid.New(),
		AccountID:     reqBody.ToAccountID,
		UserID:        userID, // Set authenticated user ID
		Credit:        reqBody.Amount,
		Debit:         "0.00",
		TransactionID: trxID,
		OperationType: "TRANSFER_IN",
		CreatedAt:     time.Now(),
	}
	if err := s.entryRepo.InsertEntry(tx, entryTo); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if err := s.accRepo.UpdateBalance(tx, reqBody.ToAccountID, reqBody.Amount, true); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error": "Failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transaction_id": trxID,
		"status":         "success",
	})
}
