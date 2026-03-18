package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerResendWebhookHandler registers a webhook endpoint handler for inbound events from Resend.
func registerResendWebhookHandler(router *Router) error {
	config := Config{
		Path:        "/resend/webhook",
		Method:      http.MethodPost,
		Name:        "ResendWebhook",
		Description: "Handle incoming webhook events from Resend for email delivery tracking",
		Tags:        []string{"webhooks", "email"},
		OperationID: "ResendWebhook",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.ResendWebhookHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}
