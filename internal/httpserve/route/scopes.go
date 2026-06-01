package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerScopesHandler registers the /scopes handler
func registerScopesHandler(router *Router) error {
	config := Config{
		Path:        "/scopes",
		Method:      http.MethodGet,
		Name:        "Scopes",
		Description: "Retrieve a list of scopes that can be configured for an api token",
		Tags:        []string{"scopes", "tokens"},
		OperationID: "Scopes",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.ScopesHandler,
	}

	return router.AddV1HandlerRoute(config)
}
