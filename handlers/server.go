package handlers

import "ledger-system/repository"

type Server struct {
	repo *repository.AccountRepository
}

func NewServer(repo *repository.AccountRepository) *Server {
	return &Server{repo: repo}
}
