package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// Login is oriented towards human users who use their email and password for
// authentication - see the handlers/login.go for more information
func registerLoginHandler(router *Router) error {
	config := Config{
		Path:        "/login",
		Method:      http.MethodPost,
		Name:        "Login",
		Description: "Login is oriented towards human users who use their email and password for authentication. Login verifies the password submitted for the user is correct by looking up the user by email and using the argon2 derived key verification process to confirm the password matches. Upon authentication an access token and a refresh token with the authorized claims of the user are returned. The user can use the access token to authenticate to our systems. The access token has an expiration and the refresh token can be used with the refresh endpoint to get a new access token without the user having to log in again. The refresh token overlaps with the access token to provide a seamless authentication experience and the user can refresh their access token so long as the refresh token is valid",
		Tags:        []string{"authentication"},
		OperationID: "LoginHandler",
		Security:    handlers.BasicSecurity(),
		Middlewares: *unauthenticatedEndpoint, // leaves off the additional middleware (including csrf)
		Handler:     router.Handler.LoginHandler,
	}

	return router.AddV1HandlerRoute(config)
}
