package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

func registerOnboardingQuestionsHandler(router *Router) error {
	config := Config{
		Path:        "/onboarding/questions",
		Method:      http.MethodGet,
		Name:        "ListOnboardingQuestions",
		Description: "List backend-defined onboarding steps and questions",
		Tags:        []string{"onboarding"},
		OperationID: "ListOnboardingQuestions",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.ListOnboardingQuestions,
	}

	return router.AddV1HandlerRoute(config)
}
