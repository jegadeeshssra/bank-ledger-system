package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		Owner    string `json:"owner"`
		Balance  int64  `json:"balance"`
		Currency string `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	acc, err := s.repo.CreateAccount(reqBody.Owner, reqBody.Balance, reqBody.Currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acc)
}
