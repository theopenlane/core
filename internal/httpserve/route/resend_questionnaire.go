package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerResendQuestionnaireHandler registers the resend questionnaire email handler and route
func registerResendQuestionnaireHandler(router *Router) error {
	config := Config{
		Path:        "/questionnaire/resend",
		Method:      http.MethodPost,
		Name:        "ResendQuestionnaireEmail",
		Description: "Resend questionnaire authentication email with new JWT token",
		Tags:        []string{"questionnaire"},
		OperationID: "ResendQuestionnaireEmail",
		Security:    handlers.BasicSecurity(),
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.ResendQuestionnaireEmail,
	}

	return router.AddUnversionedHandlerRoute(config)
}
