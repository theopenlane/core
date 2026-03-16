package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerIntegrationWebhookHandler registers the generic integration webhook handler
func registerIntegrationWebhookHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/webhook/:integrationID",
		Method:      http.MethodPost,
		Name:        "IntegrationWebhook",
		Description: "Handle one integration webhook delivery",
		Tags:        []string{"webhooks", "integrations"},
		OperationID: "IntegrationWebhook",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.IntegrationWebhookHandler,
	}

	return router.AddV1HandlerRoute(config)
}
