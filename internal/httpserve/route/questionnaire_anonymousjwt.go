package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerQuestionnaireAnonymousJWTHandler registers the questionnaire anonymous JWT handler and route
func registerQuestionnaireAnonymousJWTHandler(router *Router) error {
	config := Config{
		Path:        "/questionnaire/auth/anonymous",
		Method:      http.MethodPost,
		Name:        "QuestionnaireAnonymousJWT",
		Description: "Create anonymous JWT token for questionnaire access",
		Tags:        []string{"questionnaire", "authentication"},
		OperationID: "QuestionnaireAnonymousJWT",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.CreateQuestionnaireAnonymousJWT,
	}

	return router.AddUnversionedHandlerRoute(config)
}
