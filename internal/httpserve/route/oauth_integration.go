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
		Middlewares: authMW,
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
		Middlewares: mw, // needs to not be authenticated as it is called by external oauth provider
		Handler: func(c echo.Context) error {
			return router.Handler.HandleOAuthCallback(c)
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
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.RefreshIntegrationTokenHandler(c)
		},
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}
