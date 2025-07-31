package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerSwitchRoute registers the switch route to switch the user's logged in organization context
func registerSwitchRoute(router *Router) error {
	config := Config{
		Path:        "/switch",
		Method:      http.MethodPost,
		Name:        "Switch",
		Description: "Switch the user's organization context",
		Tags:        []string{"organization"},
		OperationID: "Switch",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *AuthenticatedEndpoint,
		Handler:     router.Handler.SwitchHandler,
	}

	return router.AddV1HandlerRoute(config)
}
