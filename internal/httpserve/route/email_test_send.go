package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerEmailTestSendHandler registers the test email send endpoint.
// Only active when the server is in dev mode and integrations are enabled
func registerEmailTestSendHandler(router *Router) error {
	if !router.Handler.IsDev || !integrationsEnabled(router) {
		return nil
	}

	config := Config{
		Path:        "/email-test/send",
		Method:      http.MethodPost,
		Name:        "EmailTestSend",
		Description: "Send test emails through registered dispatchers using scaffolded fixture data (dev mode only)",
		Tags:        []string{"email", "testing"},
		OperationID: "EmailTestSend",
		Security:    handlers.PublicSecurity,
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.EmailTestSendHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}
