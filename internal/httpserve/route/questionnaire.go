package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerQuestionnaireHandler registers the questionnaire handler and route
func registerQuestionnaireHandler(router *Router) error {
	config := Config{
		Path:         "/questionnaire",
		Method:       http.MethodGet,
		Name:         "GetQuestionnaire",
		Description:  "Get questionnaire template configuration for authenticated anonymous users",
		Tags:         []string{"Questionnaires"},
		OperationID:  "GetQuestionnaire",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.GetQuestionnaire,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerQuestionnaireSubmitHandler registers the questionnaire submit handler and route
func registerQuestionnaireSubmitHandler(router *Router) error {
	config := Config{
		Path:         "/questionnaire",
		Method:       http.MethodPost,
		Name:         "SubmitQuestionnaire",
		Description:  "Submit questionnaire response data for authenticated anonymous users",
		Tags:         []string{"Questionnaires"},
		OperationID:  "SubmitQuestionnaire",
		IncludeInOAS: true,
		Security:     handlers.AuthenticatedSecurity,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.SubmitQuestionnaire,
	}

	return router.AddUnversionedHandlerRoute(config)
}
