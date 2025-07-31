package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerAccountRolesOrganizationHandler registers the /account/roles/organization handler
func registerAccountRolesOrganizationHandler(router *Router) error {
	// add route without the path param
	config := Config{
		Path:        "/account/roles/organization",
		Method:      http.MethodGet,
		Name:        "AccountRolesOrganization",
		Description: "Retrieve a list of roles of the subject in the organization",
		Tags:        []string{"account"},
		OperationID: "AccountRolesOrganization",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *AuthenticatedEndpoint,
		Handler:     router.Handler.AccountRolesOrganizationHandler,
	}

	if err := router.AddV1HandlerRoute(config); err != nil {
		return err
	}

	// add an additional route with the path param
	configByID := Config{
		Path:        "/account/roles/organization/:id",
		Method:      http.MethodGet,
		Name:        "AccountRolesOrganizationByID",
		Description: "Retrieve a list of roles of the subject in the organization ID provided",
		Tags:        []string{"account"},
		OperationID: "AccountRolesOrganizationByID",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *AuthenticatedEndpoint,
		Handler:     router.Handler.AccountRolesOrganizationHandler,
	}

	return router.AddV1HandlerRoute(configByID)
}
