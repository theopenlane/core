package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerRegisterHandler registers the register handler and route
func registerRegisterHandler(router *Router) error {
	config := Config{
		Path:        "/register",
		Method:      http.MethodPost,
		Name:        "Register",
		Description: "Register creates a new user in the database with the specified password, allowing the user to login to Openlane. This endpoint requires a 'strong' password and a valid register request, otherwise a 400 reply is returned. The password is stored in the database as an argon2 derived key so it is impossible for a hacker to get access to raw passwords. A personal organization is created for the user registering based on the organization data in the register request and the user is assigned the Owner role",
		Tags:        []string{"accountRegistration"},
		OperationID: "Register",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.RegisterHandler,
	}

	return router.AddV1HandlerRoute(config)
}
