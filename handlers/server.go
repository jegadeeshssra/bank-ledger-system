package handlers

import (
	"database/sql"
	"ledger-system/repository"
)

type Server struct {
	accRepo   *repository.AccountRepository
	entryRepo *repository.EntryRepository
	userRepo  *repository.UserRepository
	db        *sql.DB
}

func NewServer(accRepo *repository.AccountRepository, entryRepo *repository.EntryRepository, userRepo *repository.UserRepository, db *sql.DB) *Server {
	return &Server{
		accRepo:   accRepo,
		entryRepo: entryRepo,
		userRepo:  userRepo,
		db:        db,
	}
}
