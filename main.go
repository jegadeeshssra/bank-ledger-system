package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"ledger-system/config"
	"ledger-system/db"
	"ledger-system/handlers"
	"ledger-system/middleware"
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

	// 6. Start the server using PORT/BACKEND_PORT from env
	port := config.GetString("PORT", "")
	if port == "" {
		port = config.GetString("BACKEND_PORT", "443")
	}
	port = strings.TrimPrefix(strings.TrimSpace(port), ":")
	if port == "" {
		port = "443"
	}
	listenAddr := ":" + port

	fmt.Printf("Ledger API Server starting on https://server:%s\n", port)
	handler := middleware.CORSMiddleware(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(listenAddr, handler))
}
