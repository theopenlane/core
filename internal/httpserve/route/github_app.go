package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerGitHubAppWebhookHandler registers the shared GitHub App webhook handler
func registerGitHubAppWebhookHandler(router *Router) error {
	if !integrationsEnabled(router) {
		return nil
	}

	config := Config{
		Path:        "/github/app/webhook",
		Method:      http.MethodPost,
		Name:        "GitHubAppWebhook",
		Description: "Handle GitHub App security alert webhooks",
		Tags:        []string{"webhooks", "integrations"},
		OperationID: "GitHubAppWebhook",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.GitHubAppWebhookHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}
