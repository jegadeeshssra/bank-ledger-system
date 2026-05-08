package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"ledger-system/repository"

	"github.com/google/uuid"
)

func (s *Server) GetAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	acc, err := s.accRepo.GetAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	fmt.Println(reflect.TypeOf(acc))
	if err := ValidateResponse(w, acc); err != nil {
		return
	}
	json.NewEncoder(w).Encode(acc)
}

func (s *Server) ListAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	accounts, err := s.accRepo.ListAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if accounts == nil {
		accounts = []repository.Account{}
	}
	if err := ValidateResponse(w, accounts); err != nil {
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

func (s *Server) GetEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	entries, err := s.entryRepo.GetEntriesByAccountID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []repository.Entry{}
	}
	if err := ValidateResponse(w, entries); err != nil {
		return
	}
	json.NewEncoder(w).Encode(entries)
}

// The term reconcile means to make two things consistent that should be equal but might have drifted apart.
func (s *Server) Reconcile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	balance, err := s.accRepo.CalculateBalance(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.accRepo.SetBalance(id, balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Reconciled", "balance": balance})
}

func (s *Server) GetTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	trxStr := r.PathValue("transaction_id")
	trxID, err := uuid.Parse(trxStr)
	if err != nil {
		http.Error(w, "Invalid transaction id", http.StatusBadRequest)
		return
	}

	entries, err := s.entryRepo.GetEntriesByTransactionAndAccountID(trxID, accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(entries) == 0 {
		http.Error(w, "Transaction not found or does not belong to this account", http.StatusNotFound)
		return
	}
	if err := ValidateResponse(w, entries); err != nil {
		return
	}

	json.NewEncoder(w).Encode(entries)
}
