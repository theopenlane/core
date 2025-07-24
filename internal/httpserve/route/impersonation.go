package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerStartImpersonationHandler registers the start impersonation handler
func registerStartImpersonationHandler(router *Router) (err error) {
	path := "/impersonation/start"
	method := http.MethodPost
	name := "StartImpersonation"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.StartImpersonation(c)
		},
	}

	startImpersonationOperation := router.Handler.BindStartImpersonationHandler()

	if err := router.AddV1Route(path, method, startImpersonationOperation, route); err != nil {
		return err
	}

	return nil
}

// registerEndImpersonationHandler registers the end impersonation handler
func registerEndImpersonationHandler(router *Router) (err error) {
	path := "/impersonation/end"
	method := http.MethodPost
	name := "EndImpersonation"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.EndImpersonation(c)
		},
	}

	endImpersonationOperation := router.Handler.BindEndImpersonationHandler()

	if err := router.AddV1Route(path, method, endImpersonationOperation, route); err != nil {
		return err
	}

	return nil
}
