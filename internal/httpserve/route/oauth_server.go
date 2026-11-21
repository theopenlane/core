package route

import (
	"net/http"

	iamoauth "github.com/theopenlane/iam/oauthserver"

	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
)

// registerOAuthServerHandlers registers every OAuth authorization server route, only when the server is enabled
func registerOAuthServerHandlers(router *Router) error {
	if !router.Handler.OAuthServerConfig.Enabled {
		return nil
	}

	for _, register := range []func(*Router) error{
		registerOAuthServerMetadataHandler,
		registerOAuthAuthorizeHandler,
		registerOAuthTokenHandler,
		registerOAuthRevokeHandler,
		registerOAuthAuthorizationRequestHandler,
		registerOAuthConsentHandler,
	} {
		if err := register(router); err != nil {
			return err
		}
	}

	return nil
}

// registerOAuthServerMetadataHandler serves RFC 8414 authorization server metadata
func registerOAuthServerMetadataHandler(router *Router) error {
	config := Config{
		Path:        iamoauth.MetadataPath,
		Method:      http.MethodGet,
		Name:        "OAuthServerMetadata",
		Description: "OAuth 2.0 authorization server metadata (RFC 8414)",
		Tags:        []string{"well-known", "oauth2"},
		OperationID: "OAuthServerMetadata",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		RateLimit:   publicStaticRateLimit,
		PublicCORS:  true,
		Handler:     router.Handler.OAuthServerMetadataHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerOAuthAuthorizeHandler registers the authorization endpoint
func registerOAuthAuthorizeHandler(router *Router) error {
	config := Config{
		Path:        iamoauth.AuthorizePath,
		Method:      http.MethodGet,
		Name:        "OAuthAuthorize",
		Description: "OAuth 2.1 authorization endpoint that redirects the user to the consent screen",
		Tags:        []string{"oauth2"},
		OperationID: "OAuthAuthorize",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		RateLimit:   oauthRateLimit,
		Handler:     router.Handler.OAuthAuthorizeHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerOAuthTokenHandler registers the token endpoint, with no csrf or session middleware since PKCE authenticates the exchange
func registerOAuthTokenHandler(router *Router) error {
	config := Config{
		Path:        iamoauth.TokenPath,
		Method:      http.MethodPost,
		Name:        "OAuthToken",
		Description: "OAuth 2.1 token endpoint for authorization_code and refresh_token grants",
		Tags:        []string{"oauth2"},
		OperationID: "OAuthToken",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   oauthRateLimit,
		PublicCORS:  true,
		Handler:     router.Handler.OAuthTokenHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerOAuthRevokeHandler registers the RFC 7009 revocation endpoint
func registerOAuthRevokeHandler(router *Router) error {
	config := Config{
		Path:        iamoauth.RevokePath,
		Method:      http.MethodPost,
		Name:        "OAuthRevoke",
		Description: "OAuth 2.0 token revocation endpoint (RFC 7009)",
		Tags:        []string{"oauth2"},
		OperationID: "OAuthRevoke",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   oauthRateLimit,
		PublicCORS:  true,
		Handler:     router.Handler.OAuthRevokeHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerOAuthAuthorizationRequestHandler lets the console load a pending request for the consent screen
func registerOAuthAuthorizationRequestHandler(router *Router) error {
	config := Config{
		Path:        "/oauth2/requests/:requestID",
		Method:      http.MethodGet,
		Name:        "OAuthAuthorizationRequest",
		Description: "Get a pending OAuth authorization request to render the consent screen",
		Tags:        []string{"oauth2"},
		OperationID: "OAuthAuthorizationRequest",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.OAuthAuthorizationRequestHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerOAuthConsentHandler records the user's approve or deny decision
func registerOAuthConsentHandler(router *Router) error {
	config := Config{
		Path:        "/oauth2/consent",
		Method:      http.MethodPost,
		Name:        "OAuthConsent",
		Description: "Approve or deny an OAuth authorization request and install the application in an organization",
		Tags:        []string{"oauth2"},
		OperationID: "OAuthConsent",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.OAuthConsentHandler,
	}

	return router.AddV1HandlerRoute(config)
}
