package route

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// registerVerifyHandler registers the verify handler and route which handles email verification
func registerVerifyHandler(router *Router) error {
	config := Config{
		Path:        "/verify",
		Method:      http.MethodGet,
		Name:        "VerifyEmail",
		Description: "Used to verify a user's email address - once clicked they will be redirected to the UI with a success or failure message",
		Tags:        []string{"email"},
		OperationID: "VerifyEmail",
		Security:    &openapi3.SecurityRequirements{},
		Middlewares: *RestrictedEndpoint,
		Handler:     router.Handler.VerifyEmail,
	}

	return router.AddV1HandlerRoute(config)
}
