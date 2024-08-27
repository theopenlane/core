package route

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	echo "github.com/theopenlane/echox"
)

// registerLivenessHandler registers the liveness handler
func registerLivenessHandler(router *Router) (err error) {
	path := "/livez"
	method := http.MethodGet

	route := echo.Route{
		Name:   "Livez",
		Method: method,
		Path:   path,
		Handler: func(c echo.Context) error {
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
		Name:   "Ready",
		Method: method,
		Path:   path,
		Handler: func(c echo.Context) error {
			return router.Handler.ReadyChecks.ReadyHandler(c)
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerMetricsHandler registers the metrics handler
func registerMetricsHandler(router *Router) (err error) {
	path := "/metrics"
	method := http.MethodGet

	route := echo.Route{
		Name:    "Metrics",
		Method:  method,
		Path:    path,
		Handler: echo.WrapHandler(promhttp.Handler()),
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}
