package handlers

import (
	"net/http"
	"strconv"
)

func (s *Server) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	err = s.repo.DeleteAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
