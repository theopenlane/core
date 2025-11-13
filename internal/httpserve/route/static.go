package route

import (
	"embed"
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/specs"
)

// registerOpenAPIHandler embeds our generated open api specs and serves it behind /api-docs
func registerOpenAPIHandler(router *Router) (err error) {
	config := Config{
		Path:        "/api-docs",
		Method:      http.MethodGet,
		Name:        "APIDocs",
		Description: "Get OpenAPI 3.1.1 specification for this API",
		Tags:        []string{"documentation"},
		OperationID: "APIDocs",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			if len(specs.OpenlaneSpec) == 0 {
				return ctx.NoContent(http.StatusInternalServerError)
			}

			ctx.Response().Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSON)
			return ctx.Blob(http.StatusOK, httpsling.ContentTypeJSON, specs.OpenlaneSpec)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

//go:embed robots.txt
var robotsTxt embed.FS

// registerRobotsHandler serves up the robots.txt file via the RobotsHandler
func registerRobotsHandler(router *Router) (err error) {
	config := Config{
		Path:        "/robots.txt",
		Method:      http.MethodGet,
		Name:        "Robots",
		Description: "Get robots.txt file for web crawlers",
		Tags:        []string{"static"},
		OperationID: "Robots",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return echo.StaticFileHandler("robots.txt", robotsTxt)(ctx)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

//go:embed assets/*
var assets embed.FS

// registerFaviconHandler serves up the favicon.ico
func registerFaviconHandler(router *Router) (err error) {
	config := Config{
		Path:        "/favicon.ico",
		Method:      http.MethodGet,
		Name:        "Favicon",
		Description: "Get favicon.ico for the website",
		Tags:        []string{"static"},
		OperationID: "Favicon",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return echo.StaticFileHandler("assets/favicon.ico", assets)(ctx)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerExampleCSVHandler serves up the text output of the example csv file
func registerExampleCSVHandler(router *Router) (err error) {
	config := Config{
		Path:        "/example/csv",
		Method:      http.MethodPost,
		Name:        "ExampleCSV",
		Description: "Generate and return an example CSV file for data import templates",
		Tags:        []string{"files", "examples"},
		OperationID: "ExampleCSV",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler: func(ctx echo.Context, openapi *handlers.OpenAPIContext) error {
			ctx.Response().Header().Set(httpsling.HeaderContentType, "text/csv")
			return router.Handler.ExampleCSV(ctx, openapi)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}
