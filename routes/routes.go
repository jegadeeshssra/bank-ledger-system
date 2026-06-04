package routes

import (
	"ledger-system/handlers"
)

// RegisterAllRoutes registers all API versions and routes
// Currently supporting API v1
// Future versions can be added by creating v2_auth_routes.go, v2_account_routes.go, etc.
func RegisterAllRoutes(srv *handlers.Server, authSrv *handlers.AuthServer) {
	// Register API v1 (current stable version)
	RegisterAuthRoutes(authSrv)
	RegisterAccountRoutes(srv)

	// TODO: Register API v2 when needed
	// RegisterV2AuthRoutes(authSrv)
	// RegisterV2AccountRoutes(srv)
}
