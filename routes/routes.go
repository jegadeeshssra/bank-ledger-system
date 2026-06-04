package routes

import (
	"ledger-system/handlers"
)

// RegisterAllRoutes registers all application routes
func RegisterAllRoutes(srv *handlers.Server, authSrv *handlers.AuthServer) {
	RegisterAuthRoutes(authSrv)
	RegisterAccountRoutes(srv)
}
