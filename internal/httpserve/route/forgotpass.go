package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerForgotPasswordHandler registers the forgot password handler and route
func registerForgotPasswordHandler(router *Router) error {
	config := Config{
		Path:        "/forgot-password",
		Method:      http.MethodPost,
		Name:        "ForgotPassword",
		Description: "ForgotPassword is a service for users to request a password reset email. The email address must be provided in the POST request and the user must exist in the database. This endpoint always returns 200 regardless of whether the user exists or not to avoid leaking information about users in the database",
		Tags:        []string{"forgotpassword"},
		OperationID: "ForgotPassword",
		Security:    handlers.PublicSecurity,
		Middlewares: *restrictedEndpoint,
		Handler:     router.Handler.ForgotPassword,
	}

	return router.AddV1HandlerRoute(config)
}
