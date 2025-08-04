package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerWebauthnRegistrationHandler registers the webauthn registration handler
func registerWebauthnRegistrationHandler(router *Router) (err error) {
	config := Config{
		Path:        "/registration/options",
		Method:      http.MethodPost,
		Name:        "WebauthnRegistration",
		Description: "Begin WebAuthn registration process and return credential creation options",
		Tags:        []string{"webauthn", "authentication"},
		OperationID: "WebauthnRegistration",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.BeginWebauthnRegistration,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerWebauthnVerificationsHandler registers the webauthn registration verification handler
func registerWebauthnVerificationsHandler(router *Router) (err error) {
	config := Config{
		Path:        "/registration/verification",
		Method:      http.MethodPost,
		Name:        "WebauthnRegistrationVerification",
		Description: "Complete WebAuthn registration process by verifying the credential creation response",
		Tags:        []string{"webauthn", "authentication"},
		OperationID: "WebauthnRegistrationVerification",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.FinishWebauthnRegistration,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerWebauthnAuthenticationHandler registers the webauthn authentication handler
func registerWebauthnAuthenticationHandler(router *Router) (err error) {
	config := Config{
		Path:        "/authentication/options",
		Method:      http.MethodPost,
		Name:        "WebauthnAuthentication",
		Description: "Begin WebAuthn authentication process and return credential request options",
		Tags:        []string{"webauthn", "authentication"},
		OperationID: "WebauthnAuthentication",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.BeginWebauthnLogin,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerWebauthnAuthVerificationHandler registers the webauthn authentication verification handler
func registerWebauthnAuthVerificationHandler(router *Router) (err error) {
	config := Config{
		Path:        "/authentication/verification",
		Method:      http.MethodPost,
		Name:        "WebauthnAuthenticationVerification",
		Description: "Complete WebAuthn authentication process by verifying the authentication response",
		Tags:        []string{"webauthn", "authentication"},
		OperationID: "WebauthnAuthenticationVerification",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.FinishWebauthnLogin,
	}

	return router.AddUnversionedHandlerRoute(config)
}
