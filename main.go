package main

import (
	"fmt"
	"log"
	"net/http"

	"ledger-system/db"
	"ledger-system/handlers"
	"ledger-system/repository"
)

func main() {
	// 1. Connect to the database using our db package
	database, err := db.Connect()
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer database.Close()
	fmt.Println("Successfully connected to the database!")

	// 2. Initialize the Account & Entry repositories
	accRepo := repository.NewAccountRepository(database)
	entryRepo := repository.NewEntryRepository(database)

	// 3. Create the tables if they don't exist
	err = accRepo.CreateTable()
	if err != nil {
		log.Fatal("Error creating accounts table:", err)
	}
	err = entryRepo.CreateTable()
	if err != nil {
		log.Fatal("Error creating entries table:", err)
	}

	// 4. Initialize HTTP Server Handlers
	srv := handlers.NewServer(accRepo, entryRepo, database)

	// Use Go 1.22+ clean method-based routing
	http.HandleFunc("GET /accounts", srv.ListAccounts)
	http.HandleFunc("POST /accounts", srv.CreateAccount)
	http.HandleFunc("GET /accounts/{id}", srv.GetAccount)
	http.HandleFunc("POST /accounts/{id}/deposit", srv.Deposit)
	http.HandleFunc("POST /accounts/{id}/withdraw", srv.Withdraw)
	http.HandleFunc("GET /accounts/{id}/entries", srv.GetEntries)
	http.HandleFunc("GET /accounts/{id}/reconcile", srv.Reconcile)
	http.HandleFunc("POST /accounts/{id}/transfers", srv.Transfer)
	http.HandleFunc("GET /accounts/{id}/transactions/{transaction-id}", srv.GetTransaction)

	// 5. Start the server on 8081 (since 8080 is used by playing-with-DB)
	fmt.Println("Ledger API Server starting on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
