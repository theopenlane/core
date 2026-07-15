package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerVerifyHandler registers the verify handler and route which handles email verification
func registerVerifyHandler(router *Router) error {
	config := Config{
		Path:         "/verify",
		Method:       http.MethodGet,
		Name:         "VerifyEmail",
		Description:  "Used to verify a user's email address - once clicked they will be redirected to the UI with a success or failure message",
		Tags:         []string{"Account Registration"},
		OperationID:  "VerifyEmail",
		IncludeInOAS: true,
		Security:     handlers.PublicSecurity,
		Middlewares:  *unauthenticatedEndpoint,
		RateLimit:    authRateLimit,
		Handler:      router.Handler.VerifyEmail,
	}

	return router.AddV1HandlerRoute(config)
}
