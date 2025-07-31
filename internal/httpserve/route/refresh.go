package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerRefreshHandler registers the refresh handler and route
func registerRefreshHandler(router *Router) error {
	config := Config{
		Path:        "/refresh",
		Method:      http.MethodPost,
		Name:        "Refresh",
		Description: "The Refresh endpoint re-authenticates users and API keys using a refresh token rather than requiring a username and password or API key credentials a second time and returns a new access and refresh token pair with the current credentials of the user. This endpoint is intended to facilitate long-running connections to the systems that last longer than the duration of an access token; e.g. long sessions on the UI or (especially) long running publishers and subscribers (machine users) that need to stay authenticated semi-permanently.",
		Tags:        []string{"refresh"},
		OperationID: "RefreshHandler",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *PublicEndpoint,
		Handler:     router.Handler.RefreshHandler,
	}

	return router.AddV1HandlerRoute(config)
}
