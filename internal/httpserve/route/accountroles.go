package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerAccountRolesHandler registers the /account/roles handler
func registerAccountRolesHandler(router *Router) error {
	config := Config{
		Path:        "/account/roles",
		Method:      http.MethodPost,
		Name:        "AccountRoles",
		Description: "Retrieve a list of roles of the subject in the organization",
		Tags:        []string{"account"},
		OperationID: "AccountRoles",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.AccountRolesHandler,
	}

	return router.AddV1HandlerRoute(config)
}
