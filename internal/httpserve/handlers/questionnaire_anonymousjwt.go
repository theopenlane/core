package handlers

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	models "github.com/theopenlane/core/pkg/openapi"
)

// CreateQuestionnaireAnonymousJWT creates an anonymous JWT token for questionnaire access
// The token is scoped to a specific assessment and organization
func (h *Handler) CreateQuestionnaireAnonymousJWT(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		response := models.CreateQuestionnaireAnonymousJWTResponse{}
		return h.Success(ctx, response, openapi)
	}

	// Get assessment ID from query parameters
	assessmentID := ctx.QueryParam("assessment_id")
	if assessmentID == "" {
		return h.BadRequest(ctx, ErrMissingAssessmentID, openapi)
	}

	reqCtx := ctx.Request().Context()
	// Allow database queries for assessment lookup without authentication
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.QuestionnaireContextKey{})

	// Query the assessment to verify it exists and get the owner ID
	assessmentEntity, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(assessmentID)).
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound, openapi)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	// Get email from query parameter (optional for initial token generation)
	email := ctx.QueryParam("email")
	
	// Generate anonymous session with the assessment ID, org ID, and email
	auth, err := h.AuthManager.GenerateAnonymousQuestionnaireSession(reqCtx, ctx.Response().Writer, assessmentEntity.OwnerID, assessmentEntity.ID, email)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	response := models.CreateQuestionnaireAnonymousJWTResponse{
		AuthData: *auth,
	}

	return h.Success(ctx, response, openapi)
}

