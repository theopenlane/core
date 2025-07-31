package route

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// registerWebfingerHandler registers the /.well-known/webfinger handler
func registerWebfingerHandler(router *Router) error {
	config := Config{
		Path:        "/.well-known/webfinger",
		Method:      http.MethodGet,
		Name:        "Webfinger",
		Description: "WebFinger endpoint for federated identity",
		Tags:        []string{"webfinger"},
		OperationID: "Webfinger",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *PublicEndpoint,
		Handler:     router.Handler.WebfingerHandler,
	}

	// unversioned because .well-known
	return router.AddUnversionedHandlerRoute(config)
}

// registerSSOLoginHandler starts the OIDC login flow.
func registerSSOLoginHandler(router *Router) error {
	config := Config{
		Path:        "/sso/login",
		Method:      http.MethodGet,
		Name:        "SSOLogin",
		Description: "Initiate SSO login flow",
		Tags:        []string{"sso"},
		OperationID: "SSOLogin",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *PublicEndpoint,
		Handler:     router.Handler.SSOLoginHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerSSOCallbackHandler completes the OIDC login flow.
func registerSSOCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/sso/callback",
		Method:      http.MethodGet,
		Name:        "SSOCallback",
		Description: "Complete SSO login flow callback",
		Tags:        []string{"sso"},
		OperationID: "SSOCallback",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *PublicEndpoint,
		Handler:     router.Handler.SSOCallbackHandler,
	}

	return router.AddV1HandlerRoute(config)
}
