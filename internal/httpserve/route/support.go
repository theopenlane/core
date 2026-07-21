package route

import (
	"net/http"
)

// registerSupportCallbackHandler registers the second factor of the Openlane support access flow,
// completing the configured identity provider exchange and minting the support session token
func registerSupportCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/support/callback",
		Method:      http.MethodPost,
		Name:        "SupportCallback",
		Description: "Second factor of Openlane support access: complete the configured identity provider exchange, enforce the domain restriction, and mint the support session token.",
		Tags:        []string{"support"},
		OperationID: "SupportCallback",
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.SupportCallbackHandler,
	}

	return router.AddV1HandlerRoute(config)
}
