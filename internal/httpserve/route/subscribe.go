package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerVerifySubscribeHandler registers the verify subscription handler and route
func registerVerifySubscribeHandler(router *Router) (err error) {
	path := "/subscribe/verify"
	method := http.MethodGet
	name := "VerifySubscription"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: restrictedEndpointsMW,
		Handler: func(c echo.Context) error {
			return router.Handler.VerifySubscriptionHandler(c)
		},
	}

	subscribeOperation := router.Handler.BindVerifySubscriberHandler()

	if err := router.AddV1Route(path, method, subscribeOperation, route); err != nil {
		return err
	}

	return nil
}
