package handlers

import (
	"encoding/json"
	"net/http"

	"ledger-system/repository"

	"github.com/google/uuid"
)

type CreateAccountReq struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
	IsSystem bool   `json:"is_system"`
}

func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody CreateAccountReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	acc := &repository.Account{
		ID:       uuid.New(),
		Name:     reqBody.Name,
		Balance:  "0.00",
		Currency: reqBody.Currency,
		IsSystem: reqBody.IsSystem,
	}

	err := s.accRepo.CreateAccount(acc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acc)
}

type AmountReq struct {
	Amount string `json:"amount"`
}

func (s *Server) Deposit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	var reqBody AmountReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	trxID := uuid.New()
	entry := &repository.Entry{
		ID:            uuid.New(),
		AccountID:     accountID,
		Credit:        reqBody.Amount,
		Debit:         "0.00",
		TransactionID: trxID,
		OperationType: "DEPOSIT",
	}

	if err := s.entryRepo.InsertEntry(tx, entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entry)
}

func (s *Server) Withdraw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	var reqBody AmountReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	trxID := uuid.New()
	entry := &repository.Entry{
		ID:            uuid.New(),
		AccountID:     accountID,
		Credit:        "0.00",
		Debit:         reqBody.Amount,
		TransactionID: trxID,
		OperationType: "WITHDRAW",
	}

	if err := s.entryRepo.InsertEntry(tx, entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.accRepo.UpdateBalance(tx, accountID, reqBody.Amount, false); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// When you call tx.Commit() in a database transaction, the transaction is permanently saved to the database.
	// After that point, a Rollback() call does nothing because there's nothing left to undo.
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entry)
}

type TransferReq struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        string    `json:"amount"`
}

func (s *Server) Transfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var reqBody TransferReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	trxID := uuid.New()

	// From Account (Debit)
	entryFrom := &repository.Entry{
		ID:            uuid.New(),
		AccountID:     reqBody.FromAccountID,
		Credit:        "0.00",
		Debit:         reqBody.Amount,
		TransactionID: trxID,
		OperationType: "TRANSFER_OUT",
	}
	if err := s.entryRepo.InsertEntry(tx, entryFrom); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.accRepo.UpdateBalance(tx, reqBody.FromAccountID, reqBody.Amount, false); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// To Account (Credit)
	entryTo := &repository.Entry{
		ID:            uuid.New(),
		AccountID:     reqBody.ToAccountID,
		Credit:        reqBody.Amount,
		Debit:         "0.00",
		TransactionID: trxID,
		OperationType: "TRANSFER_IN",
	}
	if err := s.entryRepo.InsertEntry(tx, entryTo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.accRepo.UpdateBalance(tx, reqBody.ToAccountID, reqBody.Amount, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transaction_id": trxID,
		"status":         "success",
	})
}
