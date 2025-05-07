package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerResetPasswordHandler registers the reset password handler and route
func registerResetPasswordHandler(router *Router) (err error) {
	path := "/password-reset"
	method := http.MethodPost
	name := "ResetPassword"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.ResetPassword(c)
		},
	}

	resetOperation := router.Handler.BindResetPasswordHandler()

	if err := router.AddV1Route(path, method, resetOperation, route); err != nil {
		return err
	}

	return nil
}
