package route

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// registerVerifySubscribeHandler registers the verify subscription handler and route
func registerVerifySubscribeHandler(router *Router) error {
	config := Config{
		Path:        "/subscribe/verify",
		Method:      http.MethodGet,
		Name:        "VerifySubscription",
		Description: "Verify a subscription",
		Tags:        []string{"subscription"},
		OperationID: "VerifySubscription",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *restrictedEndpoint,
		Handler:     router.Handler.VerifySubscriptionHandler,
	}

	return router.AddV1HandlerRoute(config)
}
