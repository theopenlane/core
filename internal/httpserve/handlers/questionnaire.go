package handlers

import (
	"context"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/common/enums"
	models "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
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

	assessmentResponse, err := h.DBClient.AssessmentResponse.Query().
		Where(
			assessmentresponse.AssessmentIDEQ(assessmentID),
			assessmentresponse.EmailEQ(email),
		).
		Only(allowCtx)
	if err != nil && !generated.IsNotFound(err) {
		logx.FromContext(reqCtx).Err(err).Msg("could not fetch assessment response")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if assessmentResponse != nil {
		if assessmentResponse.Status == enums.AssessmentResponseStatusCompleted {
			return h.BadRequest(ctx, ErrAssessmentResponseAlreadyCompleted, openapi)
		}

		if !assessmentResponse.DueDate.IsZero() && time.Now().After(assessmentResponse.DueDate) {
			_, err = h.DBClient.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
				SetStatus(enums.AssessmentResponseStatusOverdue).
				Save(allowCtx)
			if err != nil {
				logx.FromContext(reqCtx).Err(err).Msg("could not update assessment response due date")
				return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
			}

			return h.BadRequest(ctx, ErrAssessmentResponseOverdue, openapi)
		}
	}

	assessment, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(assessmentID)).
		WithTemplate().
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound, openapi)
		}

		logx.FromContext(reqCtx).Err(err).Msg("could not fetch assessment")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	response := models.GetQuestionnaireResponse{
		Jsonconfig: assessment.Jsonconfig,
		UISchema:   assessment.Uischema,
	}

	if assessment.Edges.Template != nil {
		response.Jsonconfig = assessment.Edges.Template.Jsonconfig
		response.UISchema = assessment.Edges.Template.Uischema
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

	// check if the user is authenticated
	au, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error getting authenticated user")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	var (
		assessmentID string
		email        string
		allowCtx     context.Context
	)

	allowCtx = privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.QuestionnaireContextKey{})

	if anonUser, ok := auth.AnonymousQuestionnaireUserFromContext(reqCtx); ok {
		assessmentID = anonUser.AssessmentID
		email = anonUser.SubjectEmail

		allowCtx = auth.WithAnonymousQuestionnaireUser(allowCtx, anonUser)

	} else {

		// for regular/normal authenticated users, we expect the assessment id to be passed
		// in the request by the client.
		//
		// for anon users, it's embedded inside the jwt already
		if req.AssessmentID == "" {
			return h.BadRequest(ctx, ErrMissingAssessmentID, openapi)
		}

		assessmentID = req.AssessmentID
		email = au.SubjectEmail
	}

	if assessmentID == "" {
		return h.BadRequest(ctx, ErrMissingAssessmentID, openapi)
	}

	if len(req.Data) == 0 {
		return h.BadRequest(ctx, ErrMissingQuestionnaireData, openapi)
	}

	assessment, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(assessmentID)).
		WithAssessmentResponses().
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound, openapi)
		}

		logx.FromContext(reqCtx).Err(err).Msg("could not fetch assessment")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if email == "" {
		return h.BadRequest(ctx, ErrMissingEmail, openapi)
	}

	assessmentResponse, err := h.DBClient.AssessmentResponse.Query().
		Where(assessmentresponse.EmailEqualFold(email),
			assessmentresponse.AssessmentIDEQ(assessmentID)).
		Only(allowCtx)

	if generated.IsNotFound(err) {
		return h.NotFound(ctx, ErrAssessmentResponseNotFound, openapi)
	}

	if err != nil {
		return h.NotFound(ctx, err, openapi)
	}

	if assessmentResponse.Status == enums.AssessmentResponseStatusCompleted {
		return h.BadRequest(ctx, ErrAssessmentResponseAlreadyCompleted, openapi)
	}

	documentDataQuery := h.DBClient.DocumentData.Create().
		SetOwnerID(assessment.OwnerID)

	if assessment.TemplateID != "" {
		documentDataQuery = documentDataQuery.SetTemplateID(assessment.TemplateID)
	}

	documentData, err := documentDataQuery.SetData(req.Data).Save(allowCtx)
	if err != nil {
		logx.FromContext(reqCtx).Err(err).Msg("could not create document data")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	completedAt := time.Now()

	freshResponse, err := h.DBClient.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetDocumentDataID(documentData.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		Save(allowCtx)
	if err != nil {
		logx.FromContext(reqCtx).Err(err).Msg("could not update assessment")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	response := models.SubmitQuestionnaireResponse{
		DocumentDataID: documentData.ID,
		Status:         freshResponse.Status.String(),
		CompletedAt:    completedAt.Format(time.RFC3339),
	}

	return h.Success(ctx, response, openapi)
}
