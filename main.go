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

	// 2. Initialize the Account repository
	accRepo := repository.NewAccountRepository(database)

	// 3. Create the table if it doesn't exist
	err = accRepo.CreateTable()
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	// 4. Initialize HTTP Server Handlers
	srv := handlers.NewServer(accRepo)

	// Use Go 1.22+ clean method-based routing
	http.HandleFunc("GET /accounts", srv.ListAccounts)
	http.HandleFunc("POST /accounts", srv.CreateAccount)
	http.HandleFunc("GET /accounts/{id}", srv.GetAccount)
	http.HandleFunc("DELETE /accounts/{id}", srv.DeleteAccount)

	// 5. Start the server on 8081 (since 8080 is used by playing-with-DB)
	fmt.Println("Ledger API Server starting on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
