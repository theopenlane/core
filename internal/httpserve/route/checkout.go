package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerCheckoutSessionHandler registers the checkout session handler
func registerCheckoutSessionHandler(router *Router) (err error) {
	path := "/checkout/session"
	method := http.MethodGet
	name := "CheckoutSession"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.CheckoutSessionHandler(c)
		},
	}

	//	uploadOperation := router.Handler.BindUploadBander()

	if err := router.Addv1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerCheckoutSuccessHandler registers the checkout success handler
func registerCheckoutSuccessHandler(router *Router) (err error) {
	path := "/checkout/success"
	method := http.MethodGet
	name := "CheckoutSuccess"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.CheckoutSuccessHandler(c)
		},
	}

	//	uploadOperation := router.Handler.BindUploadBander()

	if err := router.Addv1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}
