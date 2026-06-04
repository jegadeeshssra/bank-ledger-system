package routes

import (
	"net/http"

	"ledger-system/handlers"
	"ledger-system/middleware"
)

// RegisterAccountRoutes registers all account-related routes
func RegisterAccountRoutes(srv *handlers.Server) {
	// GET routes
	http.HandleFunc("GET /accounts", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("GET", "getaccounts")(srv.ListAccounts)))
	http.HandleFunc("GET /accounts/{id}", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("GET", "getaccount")(srv.GetAccount)))
	http.HandleFunc("GET /accounts/{id}/transactions/{transaction_id}", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("GET", "gettransaction")(srv.GetTransaction)))
	http.HandleFunc("GET /accounts/{id}/entries", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("GET", "getentries")(srv.GetEntries)))

	// PUT routes
	http.HandleFunc("PUT /accounts/{id}/reconcile", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("PUT", "reconcile")(srv.Reconcile)))

	// POST routes
	http.HandleFunc("POST /accounts", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("POST", "createaccount")(srv.CreateAccount)))
	http.HandleFunc("POST /accounts/{id}/deposit", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("POST", "deposit")(middleware.IdempotencyMiddleware(srv.Deposit))))
	http.HandleFunc("POST /accounts/{id}/withdraw", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("POST", "withdraw")(middleware.IdempotencyMiddleware(srv.Withdraw))))
	http.HandleFunc("POST /accounts/{id}/transfers", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("POST", "transfer")(middleware.IdempotencyMiddleware(srv.Transfer))))

	// DELETE routes
	http.HandleFunc("DELETE /accounts/{id}", middleware.JWTMiddleware(middleware.AuthRateLimitMiddleware("DELETE", "deleteaccount")(srv.DeleteAccount)))
}
