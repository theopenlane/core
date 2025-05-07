package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerAccountRolesHandler registers the /account/roles handler
func registerAccountRolesHandler(router *Router) (err error) {
	path := "/account/roles"
	method := http.MethodPost
	name := "AccountRoles"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.AccountRolesHandler(c)
		},
	}

	rolesOperation := router.Handler.BindAccountRoles()

	if err := router.AddV1Route(path, method, rolesOperation, route); err != nil {
		return err
	}

	return nil
}
