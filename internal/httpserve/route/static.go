package route

import (
	"embed"
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"
)

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

	if err := router.AddEchoOnlyRoute(route); err != nil {
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

	if err := router.AddEchoOnlyRoute(route); err != nil {
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

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}

// registerExampleCSVHandler serves up the text output of the example csv file
func registerExampleCSVHandler(router *Router) (err error) {
	path := "/example/csv"
	method := http.MethodPost
	name := "ExampleCSV"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			c.Response().Header().Set(httpsling.HeaderContentType, "text/csv")

			return router.Handler.ExampleCSV(c)
		},
	}

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}
