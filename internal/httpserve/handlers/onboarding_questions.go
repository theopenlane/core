package handlers

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/onboarding"
	"github.com/theopenlane/core/pkg/logx"
)

func (h *Handler) ListOnboardingQuestions(ctx echo.Context) error {
	if _, err := BindAndValidate[models.OnboardingQuestionsRequest](ctx); err != nil {
		return h.InvalidInput(ctx, err)
	}

	questionnaire, err := onboarding.Catalog(ctx.Request().Context(), h.DBClient)
	if err != nil {
		logx.FromContext(ctx.Request().Context()).Error().Err(err).Msg("unable to build onboarding questionnaire")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, models.OnboardingQuestionsResponse{
		Reply:   rout.Reply{Success: true},
		Version: questionnaire.Version,
		Steps:   questionnaire.Steps,
	})
}
