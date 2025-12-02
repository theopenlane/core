package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/ent/entdb"
)

// registerLivenessHandler registers the liveness handler
func registerLivenessHandler(router *Router) (err error) {
	config := Config{
		Path:        "/livez",
		Method:      http.MethodGet,
		Name:        "Livez",
		Description: "Health check endpoint to verify the service is alive",
		Tags:        []string{"health"},
		OperationID: "Livez",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			if entdb.IsShuttingDown() {
				return ctx.JSON(http.StatusServiceUnavailable, echo.Map{"status": "shutting down"})
			}

			return ctx.JSON(http.StatusOK, echo.Map{
				"status": "UP",
			})
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerReadinessHandler registers the readiness handler
func registerReadinessHandler(router *Router) (err error) {
	config := Config{
		Path:        "/ready",
		Method:      http.MethodGet,
		Name:        "Ready",
		Description: "Readiness check endpoint to verify the service is ready to accept traffic",
		Tags:        []string{"health"},
		OperationID: "Ready",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint, // leaves off the additional middleware(including csrf)
		SimpleHandler: func(ctx echo.Context) error {
			return router.Handler.ReadyChecks.ReadyHandler(ctx)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}
