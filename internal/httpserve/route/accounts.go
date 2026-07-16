package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerAccountAccessHandler registers the /account/access handler
func registerAccountAccessHandler(router *Router) error {
	config := Config{
		Path:         "/account/access",
		Method:       http.MethodPost,
		Name:         "AccountAccess",
		Description:  "Check Subject Access to Object",
		Tags:         []string{"Account Management"},
		OperationID:  "AccountAccess",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AccountAccessHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerAccountRolesOrganizationHandler registers the /account/roles/organization handler
func registerAccountRolesOrganizationHandler(router *Router) error {
	// add route without the path param
	config := Config{
		Path:         "/account/roles/organization",
		Method:       http.MethodGet,
		Name:         "AccountRolesOrganization",
		Description:  "Retrieve a list of roles of the subject in the organization",
		Tags:         []string{"Account Management"},
		OperationID:  "AccountRolesOrganization",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AccountRolesOrganizationHandler,
	}

	if err := router.AddV1HandlerRoute(config); err != nil {
		return err
	}

	// add an additional route with the path param
	configByID := Config{
		Path:         "/account/roles/organization/:id",
		Method:       http.MethodGet,
		Name:         "AccountRolesOrganizationByID",
		Description:  "Retrieve a list of roles of the subject in the organization ID provided",
		Tags:         []string{"Account Management"},
		OperationID:  "AccountRolesOrganizationByID",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AccountRolesOrganizationHandler,
	}

	return router.AddV1HandlerRoute(configByID)
}

// registerAccountRolesHandler registers the /account/roles handler
func registerAccountRolesHandler(router *Router) error {
	config := Config{
		Path:         "/account/roles",
		Method:       http.MethodPost,
		Name:         "AccountRoles",
		Description:  "Retrieve a list of roles of the subject in the organization",
		Tags:         []string{"Account Management"},
		OperationID:  "AccountRoles",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AccountRolesHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerAccountFeaturesHandler registers the /account/features handler
func registerAccountFeaturesHandler(router *Router) error {
	// add route without the path param
	config := Config{
		Path:         "/account/features",
		Method:       http.MethodGet,
		Name:         "AccountFeatures",
		Description:  "List features a subject has in relation to the authenticated organization",
		Tags:         []string{"Account Management"},
		OperationID:  "AccountFeatures",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AccountFeaturesHandler,
	}

	if err := router.AddV1HandlerRoute(config); err != nil {
		return err
	}

	// add an additional route with the path param
	configByID := Config{
		Path:         "/account/features/:id",
		Method:       http.MethodGet,
		Name:         "AccountFeaturesByID",
		Description:  "List the features a subject has in relation to the organization ID provided",
		Tags:         []string{"Account Management"},
		OperationID:  "AccountFeaturesByID",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AccountFeaturesHandler,
	}

	return router.AddV1HandlerRoute(configByID)
}
