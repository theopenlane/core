package route

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// registerVerifySubscribeHandler registers the verify subscription handler and route
func registerVerifySubscribeHandler(router *Router) error {
	config := Config{
		Path:        "/subscribe/verify",
		Method:      http.MethodPost,
		Name:        "VerifySubscription",
		Description: "Verify a subscription",
		Tags:        []string{"subscription"},
		OperationID: "VerifySubscription",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.VerifySubscriptionHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerUnsubscribeHandler registers the unsubscribe handler and route
func registerUnsubscribeHandler(router *Router) error {
	config := Config{
		Path:        "/unsubscribe",
		Method:      http.MethodPost,
		Name:        "Unsubscribe",
		Description: "Unsubscribe from communications",
		Tags:        []string{"subscription"},
		OperationID: "Unsubscribe",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.UnsubscribeHandler,
	}

	return router.AddV1HandlerRoute(config)
}
