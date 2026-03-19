package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerIntegrationWebhookHandler registers the generic integration webhook handler
func registerIntegrationWebhookHandler(router *Router) error {
	if !integrationsEnabled(router) {
		return nil
	}

	config := Config{
		Path:        "/integrations/webhook/:endpointID",
		Method:      http.MethodPost,
		Name:        "IntegrationWebhook",
		Description: "Handle one installation-scoped integration webhook delivery",
		Tags:        []string{"webhooks", "integrations"},
		OperationID: "IntegrationWebhook",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.IntegrationWebhookHandler,
	}

	return router.AddV1HandlerRoute(config)
}
