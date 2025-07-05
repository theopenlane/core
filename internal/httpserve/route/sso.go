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

	// unversioned because .well-known
	if err := router.AddUnversionedRoute(path, method, op, route); err != nil {
		return err
	}

	return nil
}

// registerSSOLoginHandler starts the OIDC login flow.
func registerSSOLoginHandler(router *Router) (err error) {
	path := "/sso/login"
	method := http.MethodGet
	name := "SSOLogin"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.SSOLoginHandler(c)
		},
	}

	op := router.Handler.BindSSOLoginHandler()

	if err := router.AddV1Route(path, method, op, route); err != nil {
		return err
	}

	return nil
}

// registerSSOCallbackHandler completes the OIDC login flow.
func registerSSOCallbackHandler(router *Router) (err error) {
	path := "/sso/callback"
	method := http.MethodGet
	name := "SSOCallback"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.SSOCallbackHandler(c)
		},
	}

	op := router.Handler.BindSSOCallbackHandler()

	if err := router.AddV1Route(path, method, op, route); err != nil {
		return err
	}

	return nil
}
