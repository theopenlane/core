package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerWebauthnRegistrationHandler registers the webauthn registration handler
func registerWebauthnRegistrationHandler(router *Router) (err error) {
	path := "/registration/options"
	method := http.MethodPost
	name := "WebauthnRegistration"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.BeginWebauthnRegistration(c)
		},
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}

// registerWebauthnVerificationsHandler registers the webauthn registration verification handler
func registerWebauthnVerificationsHandler(router *Router) (err error) {
	path := "/registration/verification"
	method := http.MethodPost
	name := "WebauthnRegistrationVerification"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.FinishWebauthnRegistration(c)
		},
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}

// registerWebauthnAuthenticationHandler registers the webauthn authentication handler
func registerWebauthnAuthenticationHandler(router *Router) (err error) {
	path := "/authentication/options"
	method := http.MethodPost
	name := "WebauthnAuthentication"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.BeginWebauthnLogin(c)
		},
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}

// registerWebauthnAuthVerificationHandler registers the webauthn authentication verification handler
func registerWebauthnAuthVerificationHandler(router *Router) (err error) {
	path := "/authentication/verification"
	method := http.MethodPost
	name := "WebauthnAuthenticationVerification"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.FinishWebauthnLogin(c)
		},
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}
