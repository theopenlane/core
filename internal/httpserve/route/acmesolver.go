package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerAcmeSolverHandler registers the acme solver handler and route
func registerAcmeSolverHandler(router *Router) error {
	config := Config{
		Path:        "/.well-known/acme-challenge/:path",
		Method:      http.MethodGet,
		Name:        "AcmeSolver",
		Description: "ACME challenge solver for Let's Encrypt certificate validation",
		Tags:        []string{"acme", "certificates"},
		OperationID: "AcmeSolver",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint, // leaves off the additional middleware(including csrf)
		Handler:     router.Handler.ACMESolverHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}
