package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerAccountRolesOrganizationHandler registers the /account/roles/organization handler
func registerAccountRolesOrganizationHandler(router *Router) (err error) {
	// add route without the path param
	path := "/account/roles/organization"
	method := http.MethodGet
	name := "AccountRolesOrganization"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.AccountRolesOrganizationHandler(c)
		},
	}

	rolesOrganizationOperation := router.Handler.BindAccountRolesOrganization()

	if err := router.AddV1Route(route.Path, route.Method, rolesOrganizationOperation, route); err != nil {
		return err
	}

	// add an additional route with the path param
	route.Path = "/account/roles/organization/:{id}"
	route.Name = name + "ByID"

	rolesOrganizationOperationByID := router.Handler.BindAccountRolesOrganizationByID()

	if err := router.AddV1Route(route.Path, route.Method, rolesOrganizationOperationByID, route); err != nil {
		return err
	}

	return nil
}
