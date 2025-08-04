package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

func registerJobRunnerRegistrationHandler(router *Router) error {
	config := Config{
		Path:        "/runners",
		Method:      http.MethodPost,
		Name:        "AgentNodeRegistration",
		Description: "Register a job runner node",
		Tags:        []string{"runners"},
		OperationID: "AgentNodeRegistration",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.RegisterJobRunner,
	}

	return router.AddV1HandlerRoute(config)
}
