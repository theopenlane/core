package handlers

import (
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	models "github.com/theopenlane/core/pkg/openapi"
)

// GetQuestionnaire retrieves questionnaire template configuration for authenticated anonymous users
func (h *Handler) GetQuestionnaire(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		response := models.GetQuestionnaireResponse{
			Jsonconfig: map[string]any{},
		}
		return h.Success(ctx, response, openapi)
	}

	reqCtx := ctx.Request().Context()

	anonUser, ok := auth.AnonymousQuestionnaireUserFromContext(reqCtx)
	if !ok {
		return h.Unauthorized(ctx, ErrMissingQuestionnaireContext, openapi)
	}

	assessmentID := anonUser.AssessmentID
	if assessmentID == "" {
		return h.BadRequest(ctx, ErrMissingAssessmentID, openapi)
	}

	email := anonUser.SubjectEmail
	if email == "" {
		return h.BadRequest(ctx, ErrMissingEmail, openapi)
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.QuestionnaireContextKey{})
	allowCtx = auth.WithAnonymousQuestionnaireUser(allowCtx, anonUser)

	assessmentResponseEntity, err := h.DBClient.AssessmentResponse.Query().
		Where(
			assessmentresponse.AssessmentIDEQ(assessmentID),
			assessmentresponse.EmailEQ(email),
		).
		Only(allowCtx)
	if err != nil && !generated.IsNotFound(err) {
		return h.InternalServerError(ctx, err, openapi)
	}

	if assessmentResponseEntity != nil {
		if assessmentResponseEntity.Status == enums.AssessmentResponseStatusCompleted {
			return h.BadRequest(ctx, ErrAssessmentResponseAlreadyCompleted, openapi)
		}

		if !assessmentResponseEntity.DueDate.IsZero() && time.Now().After(assessmentResponseEntity.DueDate) {
			_, err = h.DBClient.AssessmentResponse.UpdateOneID(assessmentResponseEntity.ID).
				SetStartedAt(assessmentResponseEntity.StartedAt).
				SetStatus(enums.AssessmentResponseStatusOverdue).
				Save(allowCtx)
			if err != nil {
				return h.InternalServerError(ctx, err, openapi)
			}

			return h.BadRequest(ctx, ErrAssessmentResponseOverdue, openapi)
		}
	}

	assessmentEntity, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(assessmentID)).
		WithTemplate().
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound, openapi)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	if assessmentEntity.Edges.Template == nil {
		return h.NotFound(ctx, ErrTemplateNotFound, openapi)
	}

	response := models.GetQuestionnaireResponse{
		Jsonconfig: assessmentEntity.Edges.Template.Jsonconfig,
	}

	return h.Success(ctx, response, openapi)
}

// SubmitQuestionnaire submits questionnaire response data for authenticated anonymous users
func (h *Handler) SubmitQuestionnaire(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		response := models.SubmitQuestionnaireResponse{}
		return h.Success(ctx, response, openapi)
	}

	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleSubmitQuestionnaireRequest, models.ExampleSubmitQuestionnaireResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	anonUser, ok := auth.AnonymousQuestionnaireUserFromContext(reqCtx)
	if !ok {
		return h.Unauthorized(ctx, ErrMissingQuestionnaireContext, openapi)
	}

	assessmentID := anonUser.AssessmentID
	if assessmentID == "" {
		return h.BadRequest(ctx, ErrMissingAssessmentID, openapi)
	}

	if len(req.Data) == 0 {
		return h.BadRequest(ctx, ErrMissingQuestionnaireData, openapi)
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.QuestionnaireContextKey{})
	allowCtx = auth.WithAnonymousQuestionnaireUser(allowCtx, anonUser)

	assessmentEntity, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(assessmentID)).
		WithTemplate().
		WithAssessmentResponses().
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound, openapi)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	if assessmentEntity.Edges.Template == nil {
		return h.NotFound(ctx, ErrTemplateNotFound, openapi)
	}

	email := anonUser.SubjectEmail
	if email == "" {
		return h.BadRequest(ctx, ErrMissingEmail, openapi)
	}

	var assessmentResponseEntity *generated.AssessmentResponse
	for _, ar := range assessmentEntity.Edges.AssessmentResponses {
		if ar.Email == email {
			assessmentResponseEntity = ar
			break
		}
	}

	if assessmentResponseEntity == nil {
		return h.NotFound(ctx, ErrAssessmentResponseNotFound, openapi)
	}

	if assessmentResponseEntity.Status == enums.AssessmentResponseStatusCompleted {
		return h.BadRequest(ctx, ErrAssessmentResponseAlreadyCompleted, openapi)
	}

	documentData, err := h.DBClient.DocumentData.Create().
		SetTemplateID(assessmentEntity.Edges.Template.ID).
		SetOwnerID(assessmentEntity.OwnerID).
		SetData(req.Data).
		Save(allowCtx)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	completedAt := time.Now()

	updatedResponse, err := h.DBClient.AssessmentResponse.UpdateOneID(assessmentResponseEntity.ID).
		SetDocumentDataID(documentData.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		Save(allowCtx)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	response := models.SubmitQuestionnaireResponse{
		DocumentDataID: documentData.ID,
		Status:         updatedResponse.Status.String(),
		CompletedAt:    completedAt.Format(time.RFC3339),
	}

	return h.Success(ctx, response, openapi)
}
