package route

import (
	"net/http"

	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
)

// Login is oriented towards human users who use their email and password for
// authentication - see the handlers/login.go for more information
func registerLoginHandler(router *Router) error {
	config := Config{
		Path:   "/login",
		Method: http.MethodPost,
		Name:   "Login",
		Description: `Login is oriented towards human users who use their email and password for authentication. Login verifies the password submitted for the user is correct by looking up the user by email and using the argon2 derived key verification process to confirm the password matches. Upon authentication an access token and a refresh token with the authorized claims of the user are returned. The user can use the access token to authenticate to our systems.The access token has an expiration and the refresh token can be used with the refresh endpoint to get a new access token without the user having to log in again. The refresh token overlaps with the access token to provide a seamless authentication experience and the user can refresh their access token so long as the refresh token is valid. For users with 2FA enabled either from the org or user level, a new token is not minted for them and they have to verify via an otp first
`,
		Tags:         []string{"Authentication"},
		OperationID:  "LoginHandler",
		Security:     handlers.PublicSecurity,
		Middlewares:  *unauthenticatedEndpoint, // leaves off the additional middleware (including csrf)
		RateLimit:    authRateLimit,
		IncludeInOAS: true,
		Handler:      router.Handler.LoginHandler,
	}

	return router.AddV1HandlerRoute(config)
}

func registerLogin2FAHandler(router *Router) error {
	config := Config{
		Path:        "/login/2fa/validate",
		Method:      http.MethodPost,
		Name:        "LoginTFAValidation",
		Description: "Complete a pending login by validating a TOTP or recovery code",
		Tags:        []string{"Authentication"},
		OperationID: "ValidateLoginTOTP",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		RateLimit:   authRateLimit,
		Handler:     router.Handler.ValidateLoginTOTP,
	}

	return router.AddV1HandlerRoute(config)
}
