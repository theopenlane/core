package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

func registerJobRunnerRegistrationHandler(router *Router) (err error) {
	path := "/runners"
	method := http.MethodPost
	name := "AgentNodeRegistration"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.RegisterJobRunner(c)
		},
	}

	op := router.Handler.BindRegisterRunnerNode()

	return router.AddV1Route(path, method, op, route)
}
