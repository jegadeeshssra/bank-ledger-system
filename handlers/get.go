package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ledger-system/repository"
)

func (s *Server) GetAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid id parameter (must be a valid integer ID)", http.StatusBadRequest)
		return
	}

	acc, err := s.repo.GetAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if acc == nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(acc)
}

func (s *Server) ListAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accounts, err := s.repo.ListAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if accounts == nil {
		accounts = []repository.Account{}
	}

	json.NewEncoder(w).Encode(accounts)
}
