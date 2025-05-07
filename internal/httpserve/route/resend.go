package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerResendEmailHandler registers the resend email handler and route
func registerResendEmailHandler(router *Router) (err error) {
	path := "/resend"
	method := http.MethodPost
	name := "ResendEmail"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.ResendEmail(c)
		},
	}

	resendOperation := router.Handler.BindResendEmailHandler()

	if err := router.AddV1Route(path, method, resendOperation, route); err != nil {
		return err
	}

	return nil
}
