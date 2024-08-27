package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerInviteHandler registers the invite handler
func registerInviteHandler(router *Router) (err error) {
	path := "/invite"
	method := http.MethodGet
	name := "OrganizationInviteAccept"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.OrganizationInviteAccept(c)
		},
	}

	inviteOperation := router.Handler.BindOrganizationInviteAccept()

	if err := router.Addv1Route(path, method, inviteOperation, route); err != nil {
		return err
	}

	return nil
}
