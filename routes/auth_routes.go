package routes

import (
	"net/http"

	"ledger-system/handlers"
	"ledger-system/middleware"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(authSrv *handlers.AuthServer) {
	// Public routes - no authentication required
	http.HandleFunc("POST /register", middleware.RegisterRateLimitMiddleware(authSrv.Register))
	http.HandleFunc("POST /login", middleware.LoginRateLimitMiddleware(authSrv.Login))
	http.HandleFunc("POST /refresh", authSrv.Refresh)
}
