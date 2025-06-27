package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerWebfingerHandler registers the /.well-known/webfinger handler
func registerWebfingerHandler(router *Router) (err error) {
	path := "/.well-known/webfinger"
	method := http.MethodGet
	name := "Webfinger"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.WebfingerHandler(c)
		},
	}

	op := router.Handler.BindWebfingerHandler()

	if err := router.AddV1Route(path, method, op, route); err != nil {
		return err
	}

	return nil
}
