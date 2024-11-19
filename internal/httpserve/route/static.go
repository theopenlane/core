package route

import (
	"embed"
	"net/http"

	echo "github.com/theopenlane/echox"
)

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

// registerOpenAPISpecHandler embeds our generated open api specs and serves it behind /api-docs
func registerOpenAPIHandler(router *Router) (err error) {
	path := "/api-docs"
	method := http.MethodGet
	name := "APIDocs"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: echo.HandlerFunc(func(c echo.Context) error {
			return c.JSON(http.StatusOK, router.OAS)
		}),
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
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

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}

//go:embed robots.txt
var robotsTxt embed.FS

// registerRobotsHandler serves up the robots.txt file via the RobotsHandler
func registerRobotsHandler(router *Router) (err error) {
	path := "/robots.txt"
	method := http.MethodGet
	name := "Robots"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     echo.StaticFileHandler("robots.txt", robotsTxt),
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}

//go:embed assets/*
var assets embed.FS

// registerFaviconHandler serves up the favicon.ico
func registerFaviconHandler(router *Router) (err error) {
	path := "/favicon.ico"
	method := http.MethodGet
	name := "Favicon"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     echo.StaticFileHandler("assets/favicon.ico", assets),
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
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

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}
