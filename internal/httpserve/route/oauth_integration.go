package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerIntegrationOAuthStartHandler registers the OAuth start handler for integrations
func registerIntegrationOAuthStartHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/oauth/start",
		Method:      http.MethodPost,
		Name:        "StartIntegrationOAuth",
		Description: "Start OAuth flow for integration",
		Tags:        []string{"integrations"},
		OperationID: "StartIntegrationOAuth",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.StartOAuthFlow,
	}

	return router.AddV1HandlerRoute(config)
}

// registerIntegrationOAuthCallbackHandler registers the OAuth callback handler for integrations
func registerIntegrationOAuthCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/oauth/callback",
		Method:      http.MethodGet,
		Name:        "IntegrationOAuthCallback",
		Description: "Handle OAuth callback for integration",
		Tags:        []string{"integrations"},
		OperationID: "IntegrationOAuthCallback",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.HandleOAuthCallback,
	}

	return router.AddV1HandlerRoute(config)
}

// registerRefreshIntegrationTokenHandler registers the handler to refresh integration tokens
func registerRefreshIntegrationTokenHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/:provider/refresh",
		Method:      http.MethodPost,
		Name:        "RefreshIntegrationToken",
		Description: "Refresh integration token",
		Tags:        []string{"integrations"},
		OperationID: "RefreshIntegrationToken",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.RefreshIntegrationTokenHandler,
	}

	return router.AddV1HandlerRoute(config)
}
