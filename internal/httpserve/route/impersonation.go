package route

import (
	"errors"
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

var ErrAuthenticationRequired = errors.New("authentication required")

// registerStartImpersonationHandler registers the start impersonation handler
func registerStartImpersonationHandler(router *Router) error {
	config := Config{
		Path:        "/impersonation/start",
		Method:      http.MethodPost,
		Name:        "StartImpersonation",
		Description: "Start an impersonation session to act as another user for support, administrative, or testing purposes. Requires appropriate permissions and logs all impersonation activity for audit purposes.",
		Tags:        []string{"impersonation"},
		OperationID: "StartImpersonation",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.StartImpersonation,
	}

	return router.AddV1HandlerRoute(config)
}

// registerEndImpersonationHandler registers the end impersonation handler
func registerEndImpersonationHandler(router *Router) error {
	config := Config{
		Path:        "/impersonation/end",
		Method:      http.MethodPost,
		Name:        "EndImpersonation",
		Description: "End an active impersonation session and return to normal user context. Logs the end of impersonation for audit purposes.",
		Tags:        []string{"impersonation"},
		OperationID: "EndImpersonation",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.EndImpersonation,
	}

	return router.AddV1HandlerRoute(config)
}
