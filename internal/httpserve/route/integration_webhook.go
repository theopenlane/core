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

// registerStaticWebhookRoutes registers all definition-declared static-route webhooks.
// Each definition with a StaticRoute on its webhook registration gets a dedicated
// POST endpoint at that path, handled by IntegrationStaticWebhookHandler
func registerStaticWebhookRoutes(router *Router) error {
	if !integrationsEnabled(router) {
		return nil
	}

	for _, sw := range router.Handler.IntegrationsRuntime.Registry().StaticWebhooks() {
		config := Config{
			Path:        sw.StaticRoute,
			Method:      http.MethodPost,
			Name:        sw.DefinitionID + "." + sw.WebhookName,
			Description: "Handle inbound " + sw.WebhookName + " webhook",
			Tags:        []string{"webhooks", "integrations"},
			OperationID: sw.DefinitionID + "_" + sw.WebhookName,
			Security:    handlers.PublicSecurity,
			Middlewares: *unauthenticatedEndpoint,
			Handler:     router.Handler.IntegrationStaticWebhookHandler(sw.DefinitionID, sw.WebhookName),
		}

		if err := router.AddV1HandlerRoute(config); err != nil {
			return err
		}
	}

	return nil
}
