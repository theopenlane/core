package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerVerifySubscribeHandler registers the verify subscription handler and route
func registerVerifySubscribeHandler(router *Router) error {
	config := Config{
		Path:         "/subscribe/verify",
		Method:       http.MethodPost,
		Name:         "VerifySubscription",
		Description:  "Verify a subscription",
		Tags:         []string{"Subscribers"},
		OperationID:  "VerifySubscription",
		IncludeInOAS: true,
		Security:     handlers.PublicSecurity,
		Middlewares:  *unauthenticatedEndpoint,
		RateLimit:    authFlowRateLimit,
		Handler:      router.Handler.VerifySubscriptionHandler,
		PublicCORS:   true,
	}

	return router.AddV1HandlerRoute(config)
}

// registerUnsubscribeHandler registers the unsubscribe handler and route
func registerUnsubscribeHandler(router *Router) error {
	config := Config{
		Path:         "/unsubscribe",
		Method:       http.MethodPost,
		Name:         "Unsubscribe",
		Description:  "Unsubscribe from communications",
		Tags:         []string{"Subscribers"},
		OperationID:  "Unsubscribe",
		IncludeInOAS: true,
		Security:     handlers.PublicSecurity,
		Middlewares:  *unauthenticatedEndpoint,
		RateLimit:    authFlowRateLimit,
		Handler:      router.Handler.UnsubscribeHandler,
		PublicCORS:   true,
	}

	return router.AddV1HandlerRoute(config)
}
