package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerRolesHandler registers the /roles handler
func registerRolesHandler(router *Router) error {
	config := Config{
		Path:        "/roles",
		Method:      http.MethodGet,
		Name:        "Roles",
		Description: "Retrieve a list of roles that can be assigned to a user in addition to their organization role",
		Tags:        []string{"scopes", "tokens"},
		OperationID: "Roles",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.RolesHandler,
	}

	return router.AddV1HandlerRoute(config)
}
