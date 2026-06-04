package main

import (
	"fmt"
	"log"
	"net/http"

	"ledger-system/db"
	"ledger-system/handlers"
	"ledger-system/repository"
	"ledger-system/routes"
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
	userRepo := repository.NewUserRepository(database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)

	// 3. Create the tables if they don't exist
	err = accRepo.CreateTable()
	if err != nil {
		log.Fatal("Error creating accounts table:", err)
	}
	err = entryRepo.CreateTable()
	if err != nil {
		log.Fatal("Error creating entries table:", err)
	}
	err = userRepo.CreateTable()
	if err != nil {
		log.Fatal("Error creating users table:", err)
	}
	err = refreshTokenRepo.CreateTable()
	if err != nil {
		log.Fatal("Error creating refresh_tokens table:", err)
	}

	// 4. Initialize HTTP Server Handlers
	srv := handlers.NewServer(accRepo, entryRepo, userRepo, database)
	authSrv := handlers.NewAuthServer(userRepo, refreshTokenRepo)

	// 5. Register all routes
	routes.RegisterAllRoutes(srv, authSrv)

	// 6. Start the server on 8081 (since 8080 is used by playing-with-DB)
	fmt.Println("Ledger API Server starting on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
