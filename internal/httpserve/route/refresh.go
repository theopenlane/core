package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerRefreshHandler registers the refresh handler and route
func registerRefreshHandler(router *Router) (err error) {
	path := "/refresh"
	method := http.MethodPost
	name := "Refresh"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.RefreshHandler(c)
		},
	}

	refreshOperation := router.Handler.BindRefreshHandler()

	if err := router.Addv1Route(path, method, refreshOperation, route); err != nil {
		return err
	}

	return nil
}
