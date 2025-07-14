package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerIntegrationOAuthStartHandler registers the OAuth start handler for integrations
func registerIntegrationOAuthStartHandler(router *Router) error {
	path := "/integrations/oauth/start"
	method := http.MethodPost
	name := "StartIntegrationOAuth"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW, // Require authentication
		Handler: func(c echo.Context) error {
			return router.Handler.StartOAuthFlow(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerIntegrationOAuthCallbackHandler registers the OAuth callback handler for integrations
func registerIntegrationOAuthCallbackHandler(router *Router) error {
	path := "/integrations/oauth/callback"
	method := http.MethodGet
	name := "IntegrationOAuthCallback"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW, // Use authMW
		Handler: func(c echo.Context) error {
			return router.Handler.HandleOAuthCallback(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerIntegrationTokenHandler registers the handler to retrieve integration tokens
func registerIntegrationTokenHandler(router *Router) error {
	path := "/integrations/:provider/token"
	method := http.MethodGet
	name := "GetIntegrationToken"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW, // Require authentication
		Handler: func(c echo.Context) error {
			return router.Handler.GetIntegrationToken(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerListIntegrationsHandler registers the handler to list organization integrations
func registerListIntegrationsHandler(router *Router) error {
	path := "/integrations"
	method := http.MethodGet
	name := "ListIntegrations"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW, // Require authentication
		Handler: func(c echo.Context) error {
			return router.Handler.ListIntegrations(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerDeleteIntegrationHandler registers the handler to delete an integration
func registerDeleteIntegrationHandler(router *Router) error {
	path := "/integrations/:id"
	method := http.MethodDelete
	name := "DeleteIntegration"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW, // Require authentication
		Handler: func(c echo.Context) error {
			return router.Handler.DeleteIntegration(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerRefreshIntegrationTokenHandler registers the handler to refresh integration tokens
func registerRefreshIntegrationTokenHandler(router *Router) error {
	path := "/integrations/:provider/refresh"
	method := http.MethodPost
	name := "RefreshIntegrationToken"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW, // Require authentication
		Handler: func(c echo.Context) error {
			return router.Handler.RefreshIntegrationTokenHandler(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerIntegrationStatusHandler registers the handler to check integration status
func registerIntegrationStatusHandler(router *Router) error {
	path := "/integrations/:provider/status"
	method := http.MethodGet
	name := "GetIntegrationStatus"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW, // Require authentication
		Handler: func(c echo.Context) error {
			return router.Handler.GetIntegrationStatus(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}
