package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerSSOTokenAuthorizeHandler registers the SSO token authorize endpoint.
func registerSSOTokenAuthorizeHandler(router *Router) error {
	path := "/sso/token/authorize"
	method := http.MethodGet
	name := "SSOTokenAuthorize"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.SSOTokenAuthorizeHandler(c)
		},
	}

	op := router.Handler.BindSSOTokenAuthorizeHandler()

	return router.AddV1Route(path, method, op, route)
}

// registerSSOTokenCallbackHandler registers the SSO token callback endpoint.
func registerSSOTokenCallbackHandler(router *Router) error {
	path := "/sso/token/callback"
	method := http.MethodGet
	name := "SSOTokenCallback"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.SSOTokenCallbackHandler(c)
		},
	}

	op := router.Handler.BindSSOTokenCallbackHandler()

	return router.AddV1Route(path, method, op, route)
}
