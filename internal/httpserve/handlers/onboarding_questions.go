package handlers

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/onboarding"
	"github.com/theopenlane/core/pkg/logx"
)

func (h *Handler) ListOnboardingQuestions(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		questionnaire, err := onboarding.Catalog(ctx.Request().Context(), nil)
		if err != nil {
			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}

		response := models.OnboardingQuestionsReply{
			Reply:   rout.Reply{Success: true},
			Version: questionnaire.Version,
			Steps:   questionnaire.Steps,
		}

		return h.Success(ctx, response, openapi)
	}

	questionnaire, err := onboarding.Catalog(ctx.Request().Context(), h.DBClient)
	if err != nil {
		logx.FromContext(ctx.Request().Context()).Error().Err(err).Msg("unable to build onboarding questionnaire")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, models.OnboardingQuestionsReply{
		Reply:   rout.Reply{Success: true},
		Version: questionnaire.Version,
		Steps:   questionnaire.Steps,
	}, openapi)
}
