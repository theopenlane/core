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
		Middlewares: *publicEndpoint,
		RateLimit:   publicStaticRateLimit,
		Handler:     router.Handler.WebfingerHandler,
	}

	// unversioned because .well-known
	return router.AddUnversionedHandlerRoute(config)
}

// registerSSOLoginHandler starts the OIDC login flow.
func registerSSOLoginHandler(router *Router) error {
	config := Config{
		Path:        "/sso/login",
		Method:      http.MethodPost,
		Name:        "SSOLogin",
		Description: "Initiate SSO login flow",
		Tags:        []string{"sso"},
		OperationID: "SSOLogin",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.SSOLoginHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerSSOInitiateHandler is the public, shareable per-organization SSO entry point.
func registerSSOInitiateHandler(router *Router) error {
	config := Config{
		Path:        "/orgs/:slug_name/sso",
		Method:      http.MethodGet,
		Name:        "SSOInitiate",
		Description: "Initiate an organization's SSO flow from its shareable slug URL",
		Tags:        []string{"sso"},
		OperationID: "SSOInitiate",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.SSOInitiateHandler,
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
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.SSOCallbackHandler,
	}

	return router.AddV1HandlerRoute(config)
}
