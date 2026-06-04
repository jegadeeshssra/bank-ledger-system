package routes

import (
	"net/http"

	"ledger-system/handlers"
	"ledger-system/middleware"
)

// RegisterAuthRoutes registers all v1 authentication-related routes
func RegisterAuthRoutes(authSrv *handlers.AuthServer) {
	// Public routes - no authentication required
	http.HandleFunc("POST /api/v1/auth/register", middleware.RegisterRateLimitMiddleware(authSrv.Register))
	http.HandleFunc("POST /api/v1/auth/login", middleware.LoginRateLimitMiddleware(authSrv.Login))
	http.HandleFunc("POST /api/v1/auth/refresh", authSrv.Refresh)
}
