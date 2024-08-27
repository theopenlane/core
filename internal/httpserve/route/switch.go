package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerSwitchRoute registers the switch route to switch the user's logged in organization context
func registerSwitchRoute(router *Router) (err error) {
	path := "/switch"
	method := http.MethodPost
	name := "Switch"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.SwitchHandler(c)
		},
	}

	switchOperation := router.Handler.BindSwitchHandler()

	if err := router.Addv1Route(path, method, switchOperation, route); err != nil {
		return err
	}

	return nil
}
