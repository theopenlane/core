package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	models "github.com/theopenlane/core/pkg/openapi"
)

func (h *Handler) CreateQuestionnaireAnonymousJWT(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		response := models.CreateQuestionnaireAnonymousJWTResponse{}
		return h.Success(ctx, response, openapi)
	}

	r, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleQuestionnaireAnonymousJWTRequest, models.ExampleCreateQuestionnaireAnonymousJWTResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	assessmentEntity, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(r.AssessmentID)).
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	auth, err := h.AuthManager.GenerateAnonymousQuestionnaireSession(reqCtx, ctx.Response().Writer, assessmentEntity.OwnerID, assessmentEntity.ID)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	return h.Success(ctx, models.CreateQuestionnaireAnonymousJWTResponse{
		AuthData: *auth,
	}, openapi)
}
