package handlers

import (
	"net/http"

	"github.com/google/uuid"
)

func (s *Server) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	err = s.accRepo.DeleteAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
