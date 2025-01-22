package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// register2faHandler registers the 2FA validation handler
// which is used to verify the TOTP code of a user
func register2faHandler(router *Router) (err error) {
	path := "/2fa/validate"
	method := http.MethodPost
	name := "2FA Validate"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.ValidateTOTP(c)
		},
	}

	tfaOperation := router.Handler.BindTFAHandler()

	if err := router.Addv1Route(path, method, tfaOperation, route); err != nil {
		return err
	}

	return nil
}
