package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerAcmeSolverHandler registers the acme solver handler and route
func registerAcmeSolverHandler(router *Router) (err error) {
	path := "/.well-known/acme-challenge/:path"
	method := http.MethodGet
	name := "AcmeSolver"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.ACMESolverHandler(c)
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}
