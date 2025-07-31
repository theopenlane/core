package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerSSOTokenAuthorizeHandler registers the SSO token authorize endpoint.
func registerSSOTokenAuthorizeHandler(router *Router) error {
	config := Config{
		Path:        "/sso/token/authorize",
		Method:      http.MethodGet,
		Name:        "SSOTokenAuthorize",
		Description: "Authorize SSO token request",
		Tags:        []string{"sso"},
		OperationID: "SSOTokenAuthorize",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *AuthenticatedEndpoint,
		Handler:     router.Handler.SSOTokenAuthorizeHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerSSOTokenCallbackHandler registers the SSO token callback endpoint.
func registerSSOTokenCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/sso/token/callback",
		Method:      http.MethodGet,
		Name:        "SSOTokenCallback",
		Description: "Handle SSO token callback",
		Tags:        []string{"sso"},
		OperationID: "SSOTokenCallback",
		Security:    handlers.PublicSecurity,
		Middlewares: *PublicEndpoint,
		Handler:     router.Handler.SSOTokenCallbackHandler,
	}

	return router.AddV1HandlerRoute(config)
}
