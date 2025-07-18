package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/entdb"
	echo "github.com/theopenlane/echox"
)

// registerLivenessHandler registers the liveness handler
func registerLivenessHandler(router *Router) (err error) {
	path := "/livez"
	method := http.MethodGet

	route := echo.Route{
		Name:        "Livez",
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			if entdb.IsShuttingDown() {
				return c.JSON(http.StatusServiceUnavailable, echo.Map{"status": "shutting down"})
			}

			return c.JSON(http.StatusOK, echo.Map{
				"status": "UP",
			})
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerReadinessHandler registers the readiness handler
func registerReadinessHandler(router *Router) (err error) {
	path := "/ready"
	method := http.MethodGet

	route := echo.Route{
		Name:        "Ready",
		Method:      method,
		Path:        path,
		Middlewares: baseMW, // leaves off the additional middleware(including csrf)
		Handler: func(c echo.Context) error {
			return router.Handler.ReadyChecks.ReadyHandler(c)
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}
