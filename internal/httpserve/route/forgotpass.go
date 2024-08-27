package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerForgotPasswordHandler registers the forgot password handler and route
func registerForgotPasswordHandler(router *Router) (err error) {
	path := "/forgot-password"
	method := http.MethodPost
	name := "ForgotPassword"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: restrictedEndpointsMW,
		Handler: func(c echo.Context) error {
			return router.Handler.ForgotPassword(c)
		},
	}

	forgotPasswordOperation := router.Handler.BindForgotPassword()

	if err := router.Addv1Route(path, method, forgotPasswordOperation, route); err != nil {
		return err
	}

	return nil
}
