package route

import (
	"net/http"

	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
)

// registerSSOLoginHandler starts the OIDC login flow.
func registerSSOLoginHandler(router *Router) error {
	config := Config{
		Path:        "/sso/login",
		Method:      http.MethodPost,
		Name:        "SSOLogin",
		Description: "Initiate SSO login flow",
		Tags:        []string{"sso"},
		OperationID: "SSOLogin",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.SSOLoginHandler,
	}

	return router.AddV1HandlerRoute(config)
}

func registerSSOTestHandler(router *Router) error {
	config := Config{
		Path:        "/sso/test",
		Method:      http.MethodPost,
		Name:        "SSOTest",
		Description: "Test and verify an already saved SSO configuration",
		Tags:        []string{"sso"},
		OperationID: "SSOVerifyConnection",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.SSOTestHandler,
	}

	return router.AddV1HandlerRoute(config)
}

func registerSSOTestCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/sso/test/callback",
		Method:      http.MethodPost,
		Name:        "SSOTestCallback",
		Description: "Complete the SSO configuration test",
		Tags:        []string{"sso"},
		OperationID: "SSOTestCallback",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.SSOTestCallbackHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerSSOInitiateHandler is the public, shareable per-organization SSO entry point.
func registerSSOInitiateHandler(router *Router) error {
	config := Config{
		Path:         "/orgs/:slug_name/sso",
		Method:       http.MethodGet,
		Name:         "SSOInitiate",
		Description:  "Initiate an organization's SSO flow from its shareable slug URL",
		Tags:         []string{"Authentication"},
		OperationID:  "SSOInitiate",
		IncludeInOAS: true,
		Security:     handlers.PublicSecurity,
		Middlewares:  *unauthenticatedEndpoint,
		RateLimit:    authFlowRateLimit,
		Handler:      router.Handler.SSOInitiateHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerSSOCallbackHandler completes the OIDC login flow.
func registerSSOCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/sso/callback",
		Method:      http.MethodPost,
		Name:        "SSOCallback",
		Description: "Complete SSO login flow callback",
		Tags:        []string{"sso"},
		OperationID: "SSOCallback",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.SSOCallbackHandler,
	}

	return router.AddV1HandlerRoute(config)
}
