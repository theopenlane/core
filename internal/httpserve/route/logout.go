package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerLogoutHandler registers the logout handler and route
func registerLogoutHandler(router *Router) error {
	config := Config{
		Path:        "/logout",
		Method:      http.MethodPost,
		Name:        "Logout",
		Description: "The Logout endpoint revokes the caller's access and refresh tokens and deletes their server-side session so that the credentials can no longer be used. It is intended to be called by clients on sign-out to ensure logout is enforced on the server rather than only clearing client-side state.",
		Tags:        []string{"logout"},
		OperationID: "LogoutHandler",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *publicEndpoint,
		RateLimit:   authFlowRateLimit,
		Handler:     router.Handler.LogoutHandler,
	}

	return router.AddV1HandlerRoute(config)
}
