package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerAccountAccessHandler registers the /account/access handler
func registerAccountAccessHandler(router *Router) (err error) {
	path := "/account/access"
	method := http.MethodPost
	name := "AccountAccess"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.AccountAccessHandler(c)
		},
	}

	accountAccessOperation := router.Handler.BindAccountAccess()

	if err := router.Addv1Route(path, method, accountAccessOperation, route); err != nil {
		return err
	}

	return nil
}
