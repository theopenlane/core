package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerIntegrationAuthStartHandler registers the auth start handler for integrations
func registerIntegrationAuthStartHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/auth/start",
		Method:      http.MethodPost,
		Name:        "StartIntegrationAuth",
		Description: "Start auth flow for integration",
		Tags:        []string{"Integrations"},
		OperationID: "StartIntegrationAuth",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.StartIntegrationAuth,
	}

	return router.AddV1HandlerRoute(config)
}

// registerIntegrationAuthCallbackHandler registers the auth callback handler for integrations
func registerIntegrationAuthCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/auth/callback",
		Method:      http.MethodGet,
		Name:        "IntegrationAuthCallback",
		Description: "Handle auth callback for integration",
		Tags:        []string{"Integrations"},
		OperationID: "IntegrationAuthCallback",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.HandleIntegrationAuthCallback,
	}

	return router.AddV1HandlerRoute(config)
}

func registerIntegrationConfigHandler(router *Router) error {
	config := Config{
		Path:         "/integrations/:definitionID/config",
		Method:       http.MethodPost,
		Name:         "ConfigureIntegrationProvider",
		Description:  "Persist integration credentials or configuration",
		Tags:         []string{"Integrations"},
		OperationID:  "ConfigureIntegrationProvider",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.ConfigureIntegrationProvider,
	}

	return router.AddV1HandlerRoute(config)
}

func registerIntegrationProvidersHandler(router *Router) error {
	config := Config{
		Path:         "/integrations/providers",
		Method:       http.MethodGet,
		Name:         "ListIntegrationProviders",
		Description:  "List available integration providers and their metadata",
		Tags:         []string{"Integrations"},
		OperationID:  "ListIntegrationProviders",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.ListIntegrationProviders,
	}

	return router.AddV1HandlerRoute(config)
}

// registerIntegrationDisconnectHandler registers the handler to disconnect an installed integration
func registerIntegrationDisconnectHandler(router *Router) error {
	config := Config{
		Path:         "/integrations/:integrationID/disconnect",
		Method:       http.MethodPost,
		Name:         "DisconnectIntegration",
		Description:  "Disconnect an installed integration and clean up credentials",
		Tags:         []string{"Integrations"},
		OperationID:  "DisconnectIntegration",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.DisconnectIntegration,
	}

	return router.AddV1HandlerRoute(config)
}

func registerIntegrationOperationHandler(router *Router) error {
	config := Config{
		Path:         "/integrations/:integrationID/operations/run",
		Method:       http.MethodPost,
		Name:         "RunIntegrationOperation",
		Description:  "Execute a provider operation using stored credentials",
		Tags:         []string{"Integrations"},
		OperationID:  "RunIntegrationOperation",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.RunIntegrationOperation,
	}

	return router.AddV1HandlerRoute(config)
}
