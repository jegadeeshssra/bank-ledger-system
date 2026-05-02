package handlers

import (
	"database/sql"
	"ledger-system/repository"
)

type Server struct {
	accRepo   *repository.AccountRepository
	entryRepo *repository.EntryRepository
	db        *sql.DB
}

func NewServer(accRepo *repository.AccountRepository, entryRepo *repository.EntryRepository, db *sql.DB) *Server {
	return &Server{
		accRepo:   accRepo,
		entryRepo: entryRepo,
		db:        db,
	}
}
