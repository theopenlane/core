package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// register2faHandler registers the 2FA validation handler which is used to verify the TOTP code of a user
func register2faHandler(router *Router) error {
	config := Config{
		Path:        "/2fa/validate",
		Method:      http.MethodPost,
		Name:        "TFAValidation",
		Description: "Validate a user's TOTP code",
		Tags:        []string{"tfa"},
		OperationID: "TFAValidation",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.ValidateTOTP,
	}

	return router.AddV1HandlerRoute(config)
}
