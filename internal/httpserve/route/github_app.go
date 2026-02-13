package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerGitHubAppInstallHandler registers the GitHub App installation start handler.
func registerGitHubAppInstallHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/github/app/install",
		Method:      http.MethodPost,
		Name:        "GitHubAppInstall",
		Description: "Start GitHub App installation flow",
		Tags:        []string{"integrations"},
		OperationID: "GitHubAppInstall",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.StartGitHubAppInstallation,
	}

	return router.AddV1HandlerRoute(config)
}

// registerGitHubAppCallbackHandler registers the GitHub App installation callback handler.
func registerGitHubAppCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/github/app/callback",
		Method:      http.MethodGet,
		Name:        "GitHubAppInstallCallback",
		Description: "Handle GitHub App installation callback",
		Tags:        []string{"integrations"},
		OperationID: "GitHubAppInstallCallback",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.GitHubAppInstallCallback,
	}

	return router.AddV1HandlerRoute(config)
}

// registerGitHubAppWebhookHandler registers the GitHub App webhook handler.
func registerGitHubAppWebhookHandler(router *Router) error {
	config := Config{
		Path:        "/github/app/webhook",
		Method:      http.MethodPost,
		Name:        "GitHubAppWebhook",
		Description: "Handle GitHub App security alert webhooks",
		Tags:        []string{"webhooks", "integrations"},
		OperationID: "GitHubAppWebhook",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.GitHubIntegrationWebhookHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}
