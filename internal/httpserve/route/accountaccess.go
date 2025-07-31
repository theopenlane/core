package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerAccountAccessHandler registers the /account/access handler
func registerAccountAccessHandler(router *Router) error {
	config := Config{
		Path:        "/account/access",
		Method:      http.MethodPost,
		Name:        "AccountAccess",
		Description: "Check Subject Access to Object",
		Tags:        []string{"account"},
		OperationID: "AccountAccess",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *AuthenticatedEndpoint,
		Handler:     router.Handler.AccountAccessHandler,
	}

	return router.AddV1HandlerRoute(config)
}
