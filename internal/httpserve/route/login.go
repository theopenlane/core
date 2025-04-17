package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// Login is oriented towards human users who use their email and password for
// authentication - see the handlers/login.go for more information
func registerLoginHandler(router *Router) (err error) {
	path := "/login"
	method := http.MethodPost
	name := "Login"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.LoginHandler(c)
		},
	}

	loginOperation := router.Handler.BindLoginHandler()

	if err := router.Addv1Route(path, method, loginOperation, route); err != nil {
		return err
	}

	return nil
}

// registerAvailableAuthHandler is oriented towards users that would love to use different
// authentication types to get into their account
func registerAvailableAuthHandler(router *Router) (err error) {
	path := "/login/methods"
	method := http.MethodPost
	name := "AvailableAuth"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.AvailableAuthTypeHandler(c)
		},
	}

	loginOperation := router.Handler.BindAvailableAuthTypeHandler()

	return router.Addv1Route(path, method, loginOperation, route)
}
