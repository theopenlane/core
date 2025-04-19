package route

import (
	"embed"
	"net/http"

	echo "github.com/theopenlane/echox"
)

//go:embed webauthn
var webauthn embed.FS

// registerSecurityTxtHandler serves up the text output of the security.txt
func registerWebAuthnWellKnownHandler(router *Router) (err error) {
	path := "/.well-known/webauthn"
	method := http.MethodGet
	name := "WebAuthn"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     echo.StaticFileHandler("webauthn", webauthn),
	}

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}

// registerJwksWellKnownHandler supplies the JWKS endpoint
func registerJwksWellKnownHandler(router *Router) (err error) {
	path := "/.well-known/jwks.json"
	method := http.MethodGet
	name := "JWKS"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return c.JSON(http.StatusOK, router.Handler.JWTKeys)
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

//go:embed security.txt
var securityTxt embed.FS

// registerSecurityTxtHandler serves up the text output of the security.txt
func registerSecurityTxtHandler(router *Router) (err error) {
	path := "/.well-known/security.txt"
	method := http.MethodGet
	name := "SecurityTxt"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     echo.StaticFileHandler("security.txt", securityTxt),
	}

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}

//go:embed applemerchant
var applemerchant embed.FS

// registerAppleMerchantHandler serves up the text output of the applemerchant file
func registerAppleMerchantHandler(router *Router) (err error) {
	path := "/.well-known/apple-developer-merchantid-domain-association"
	method := http.MethodGet
	name := "Applemerchant"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     echo.StaticFileHandler("applemerchant", applemerchant),
	}

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}
