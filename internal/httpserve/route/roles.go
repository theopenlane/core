package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerRolesHandler registers the /roles handler
func registerRolesHandler(router *Router) error {
	config := Config{
		Path:         "/roles",
		Method:       http.MethodGet,
		Name:         "Roles",
		Description:  "Retrieve a list of roles that can be assigned to a user in addition to their organization role",
		Tags:         []string{"Account Management"},
		OperationID:  "Roles",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.RolesHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerOrganizationRolesHandler registers the /account/organization-roles handler
func registerOrganizationRolesHandler(router *Router) error {
	config := Config{
		Path:         "/account/organization-roles",
		Method:       http.MethodGet,
		Name:         "OrganizationRoles",
		Description:  "Retrieve a list of organization responsibility roles that can be assigned",
		Tags:         []string{"Account Management"},
		OperationID:  "OrganizationRoles",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.RolesHandler,
	}

	if err := router.AddV1HandlerRoute(config); err != nil {
		return err
	}

	createConfig := Config{
		Path:         "/account/organization-roles",
		Method:       http.MethodPost,
		Name:         "AssignOrganizationRoles",
		Description:  "Assign an organization responsibility role to users or groups",
		Tags:         []string{"Account Management"},
		OperationID:  "AssignOrganizationRoles",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AssignOrganizationRolesHandler,
	}

	if err := router.AddV1HandlerRoute(createConfig); err != nil {
		return err
	}

	deleteConfig := Config{
		Path:         "/account/organization-roles",
		Method:       http.MethodDelete,
		Name:         "DeleteOrganizationRoles",
		Description:  "Remove an organization responsibility role from users or groups",
		Tags:         []string{"Account Management"},
		OperationID:  "DeleteOrganizationRoles",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.DeleteOrganizationRolesHandler,
	}

	return router.AddV1HandlerRoute(deleteConfig)
}

// registerAccountRolesMeHandler registers the /account/roles/me handler
func registerAccountRolesMeHandler(router *Router) error {
	config := Config{
		Path:         "/account/roles/me",
		Method:       http.MethodGet,
		Name:         "AccountRolesMe",
		Description:  "Retrieve organization responsibility roles assigned to the authenticated user",
		Tags:         []string{"Account Management"},
		OperationID:  "AccountRolesMe",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.AccountRolesMeHandler,
	}

	return router.AddV1HandlerRoute(config)
}
