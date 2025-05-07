package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerRegisterHandler registers the register handler and route
func registerRegisterHandler(router *Router) (err error) {
	path := "/register"
	method := http.MethodPost
	name := "Register"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: restrictedEndpointsMW,
		Handler: func(c echo.Context) error {
			return router.Handler.RegisterHandler(c)
		},
	}

	registerOperation := router.Handler.BindRegisterHandler()

	if err := router.AddV1Route(path, method, registerOperation, route); err != nil {
		return err
	}

	return nil
}
