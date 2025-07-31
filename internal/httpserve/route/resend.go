package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerResendEmailHandler registers the resend email handler and route
func registerResendEmailHandler(router *Router) error {
	config := Config{
		Path:        "/resend",
		Method:      http.MethodPost,
		Name:        "ResendEmail",
		Description: "Resends an email verification email to the user (only valid if the email is not already verified)",
		Tags:        []string{"resend"},
		OperationID: "ResendEmail",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.ResendEmail,
	}

	return router.AddV1HandlerRoute(config)
}
