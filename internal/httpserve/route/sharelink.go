package route

import (
	"net/http"

	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
)

// registerShareLinkResolveHandler registers the endpoint the file worker calls to describe a share link and check its
// token; it is authenticated by the forwarder secret rather than a user session
func registerShareLinkResolveHandler(router *Router) error {
	config := Config{
		Path:        "/share/:id/resolve",
		Method:      http.MethodPost,
		Name:        "ShareLinkResolve",
		Description: "Describe a share link and the files it grants downloads of",
		Tags:        []string{"files"},
		OperationID: "ShareLinkResolve",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   publicStaticRateLimit,
		Handler:     router.Handler.ShareLinkResolveHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerShareLinkCodeHandler registers the endpoint the file worker calls to email a verification code to the
// recipient of a share link; it is authenticated by the forwarder secret rather than a user session
func registerShareLinkCodeHandler(router *Router) error {
	config := Config{
		Path:        "/share/:id/code",
		Method:      http.MethodPost,
		Name:        "ShareLinkCode",
		Description: "Email a verification code to the recipient of a share link",
		Tags:        []string{"files"},
		OperationID: "ShareLinkCode",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   emailRateLimit,
		Handler:     router.Handler.ShareLinkCodeHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerShareLinkVerifyHandler registers the endpoint the file worker calls to verify the code a share link
// recipient entered; it is authenticated by the forwarder secret rather than a user session
func registerShareLinkVerifyHandler(router *Router) error {
	config := Config{
		Path:        "/share/:id/verify",
		Method:      http.MethodPost,
		Name:        "ShareLinkVerify",
		Description: "Verify the code entered by the recipient of a share link",
		Tags:        []string{"files"},
		OperationID: "ShareLinkVerify",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authRateLimit,
		Handler:     router.Handler.ShareLinkVerifyHandler,
	}

	return router.AddV1HandlerRoute(config)
}

// registerShareLinkConsumeHandler registers the endpoint the file worker calls before serving a download through a
// share link; it is authenticated by the forwarder secret rather than a user session
func registerShareLinkConsumeHandler(router *Router) error {
	config := Config{
		Path:        "/share/:id/consume",
		Method:      http.MethodPost,
		Name:        "ShareLinkConsume",
		Description: "Record a download through a share link and report whether it may be served",
		Tags:        []string{"files"},
		OperationID: "ShareLinkConsume",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   publicStaticRateLimit,
		Handler:     router.Handler.ShareLinkConsumeHandler,
	}

	return router.AddV1HandlerRoute(config)
}
