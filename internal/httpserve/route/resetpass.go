package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerResetPasswordHandler registers the reset password handler and route
func registerResetPasswordHandler(router *Router) error {
	config := Config{
		Path:        "/password-reset",
		Method:      http.MethodPost,
		Name:        "ResetPassword",
		Description: "ResetPassword allows the user (after requesting a password reset) to set a new password - the password reset token needs to be set in the request and not expired. If the request is successful, a confirmation of the reset is sent to the user and a 200 StatusOK is returned",
		Tags:        []string{"password-reset"},
		OperationID: "ResetPassword",
		Security:    handlers.PublicSecurity,
		Middlewares: *PublicEndpoint,
		Handler:     router.Handler.ResetPassword,
	}

	return router.AddV1HandlerRoute(config)
}
