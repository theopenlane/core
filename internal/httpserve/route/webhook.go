package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerWebhookHandler registers a webhook endpoint handler behind the /stripe/ path for handling inbound event receivers from stripe
func registerWebhookHandler(router *Router) (err error) {
	config := Config{
		Path:        "/stripe/webhook",
		Method:      http.MethodPost,
		Name:        "StripeWebhook",
		Description: "Handle incoming webhook events from Stripe for subscription and payment processing",
		Tags:        []string{"webhooks", "payments"},
		OperationID: "StripeWebhook",
		Security:    handlers.PublicSecurity,  // Stripe signs the request
		Middlewares: *UnauthenticatedEndpoint, // leaves off the additional middleware(including csrf)
		Handler:     router.Handler.WebhookReceiverHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}
