package handlers

import (
	"context"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
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

	assessmentID, ok := auth.ActiveAssessmentIDKey.Get(reqCtx)
	if !ok {
		return h.Unauthorized(ctx, ErrMissingQuestionnaireContext, openapi)
	}

	if assessmentID == "" {
		return h.BadRequest(ctx, ErrMissingAssessmentID, openapi)
	}

	caller, callerOk := auth.CallerFromContext(reqCtx)
	if !callerOk || caller == nil {
		return h.Unauthorized(ctx, ErrMissingQuestionnaireContext, openapi)
	}

	email := caller.SubjectEmail
	if email == "" {
		return h.BadRequest(ctx, ErrMissingEmail, openapi)
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, caller)
	allowCtx = auth.ActiveAssessmentIDKey.Set(allowCtx, assessmentID)

	assessmentResponse, err := h.DBClient.AssessmentResponse.Query().
		Where(
			assessmentresponse.AssessmentIDEQ(assessmentID),
			assessmentresponse.EmailEQ(email),
		).
		WithDocument().
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

	if assessmentResponse != nil && assessmentResponse.Edges.Document != nil {
		response.SavedData = assessmentResponse.Edges.Document.Data
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

	var (
		assessmentID string
		email        string
		allowCtx     context.Context
	)

	allowCtx = privacy.DecisionContext(reqCtx, privacy.Allow)

	if anonAssessmentID, ok := auth.ActiveAssessmentIDKey.Get(reqCtx); ok {
		assessmentID = anonAssessmentID

		anonCaller, callerOk := auth.CallerFromContext(reqCtx)
		if callerOk && anonCaller != nil {
			email = anonCaller.SubjectEmail
			allowCtx = auth.WithCaller(allowCtx, anonCaller)
		}

		allowCtx = auth.ActiveAssessmentIDKey.Set(allowCtx, assessmentID)
	} else {
		qCaller, qOk := auth.CallerFromContext(reqCtx)
		if !qOk || qCaller == nil {
			logx.FromContext(reqCtx).Error().Msg("error getting authenticated user")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}

		// for regular/normal authenticated users, we expect the assessment id to be passed
		// in the request by the client.
		//
		// for anon users, it's embedded inside the jwt already
		if req.AssessmentID == "" {
			return h.BadRequest(ctx, ErrMissingAssessmentID, openapi)
		}

		assessmentID = req.AssessmentID
		email = qCaller.SubjectEmail

		// bypass org filter and FGA tuple creation for questionnaire submissions;
		// DocumentData ownership is tracked via AssessmentResponse, not FGA tuples
		allowCtx = auth.WithCaller(allowCtx, &auth.Caller{
			OrganizationID: qCaller.OrganizationID,
			SubjectID:      qCaller.SubjectID,
			Capabilities:   auth.CapBypassFGA | auth.CapBypassOrgFilter,
		})
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

	var documentDataID string

	if assessmentResponse.DocumentDataID != "" {
		err = h.DBClient.DocumentData.UpdateOneID(assessmentResponse.DocumentDataID).
			SetData(req.Data).
			Exec(allowCtx)
		if err != nil {
			logx.FromContext(reqCtx).Err(err).Msg("could not update document data")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}

		documentDataID = assessmentResponse.DocumentDataID
	} else {
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

		documentDataID = documentData.ID
	}

	responseUpdate := h.DBClient.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetDocumentDataID(documentDataID)

	if req.IsDraft {
		responseUpdate = responseUpdate.
			SetStatus(enums.AssessmentResponseStatusDraft).
			SetIsDraft(true)
	} else {
		responseUpdate = responseUpdate.
			SetStatus(enums.AssessmentResponseStatusCompleted).
			SetCompletedAt(time.Now()).
			SetIsDraft(false)
	}

	freshResponse, err := responseUpdate.Save(allowCtx)
	if err != nil {
		logx.FromContext(reqCtx).Err(err).Msg("could not update assessment response")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	response := models.SubmitQuestionnaireResponse{
		DocumentDataID: documentDataID,
		Status:         freshResponse.Status.String(),
	}

	if !freshResponse.CompletedAt.IsZero() {
		response.CompletedAt = freshResponse.CompletedAt.Format(time.RFC3339)
	}

	return h.Success(ctx, response, openapi)
}
