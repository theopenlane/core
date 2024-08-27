package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerVerifyHandler registers the verify handler and route which handles email verification
func registerVerifyHandler(router *Router) (err error) {
	path := "/verify"
	method := http.MethodGet
	name := "VerifyEmail"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: restrictedEndpointsMW,
		Handler: func(c echo.Context) error {
			return router.Handler.VerifyEmail(c)
		},
	}

	verifyOperation := router.Handler.BindVerifyEmailHandler()

	if err := router.Addv1Route(path, method, verifyOperation, route); err != nil {
		return err
	}

	return nil
}
