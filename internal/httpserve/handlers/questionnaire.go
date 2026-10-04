package handlers

import (
	"context"
	"errors"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessment"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// GetQuestionnaire retrieves questionnaire template configuration for authenticated anonymous users
func (h *Handler) GetQuestionnaire(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()

	assessmentID, ok := auth.ActiveAssessmentIDKey.Get(reqCtx)
	if !ok {
		return h.Unauthorized(ctx, ErrMissingQuestionnaireContext)
	}

	if assessmentID == "" {
		return h.BadRequest(ctx, ErrMissingAssessmentID)
	}

	caller, callerOk := auth.CallerFromContext(reqCtx)
	if !callerOk {
		return h.Unauthorized(ctx, ErrMissingQuestionnaireContext)
	}

	email := caller.SubjectEmail

	internalCtx := auth.WithInternalOperationContext(auth.WithCaller(reqCtx, caller))
	internalCtx = auth.ActiveAssessmentIDKey.Set(internalCtx, assessmentID)

	var (
		assessmentResponse *generated.AssessmentResponse
		err                error
	)

	// email not required because we can generate anon links that can now be shared
	// but if it exists, verify that it matches the response we created when we sent it out
	// preview tokens (sender test sends) resolve to the test response; real recipients get their real response
	isPreview, _ := auth.ActiveAssessmentPreviewKey.Get(reqCtx)

	if email != "" {
		query := h.DBClient.AssessmentResponse.Query().
			Where(
				assessmentresponse.AssessmentIDEQ(assessmentID),
				assessmentresponse.EmailEQ(email),
				assessmentresponse.IsTestEQ(isPreview),
			)

		var campaignPredicate = assessmentresponse.CampaignIDIsNil()

		if id, ok := auth.ActiveCampaignIDKey.Get(reqCtx); ok && id != "" {
			campaignPredicate = assessmentresponse.CampaignIDEQ(id)
		}

		assessmentResponse, err = query.Where(campaignPredicate).
			WithDocument().
			Only(internalCtx)
		if err != nil && !generated.IsNotFound(err) {
			logx.FromContext(reqCtx).Err(err).Msg("could not fetch assessment response")
			return h.InternalServerError(ctx, ErrProcessingRequest)
		}
	}

	if assessmentResponse != nil {
		if assessmentResponse.Status == enums.AssessmentResponseStatusCompleted {
			return h.BadRequest(ctx, ErrAssessmentResponseAlreadyCompleted)
		}

		if !assessmentResponse.DueDate.IsZero() && time.Now().After(assessmentResponse.DueDate) {
			_, err = h.DBClient.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
				SetStatus(enums.AssessmentResponseStatusOverdue).
				Save(internalCtx)
			if err != nil {
				logx.FromContext(reqCtx).Err(err).Msg("could not update assessment response due date")
				return h.InternalServerError(ctx, ErrProcessingRequest)
			}

			return h.BadRequest(ctx, ErrAssessmentResponseOverdue)
		}
	}

	assessment, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(assessmentID)).
		WithTemplate().
		Only(internalCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound)
		}

		logx.FromContext(reqCtx).Err(err).Msg("could not fetch assessment")
		return h.InternalServerError(ctx, ErrProcessingRequest)
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

	return h.Success(ctx, response)
}

// SubmitQuestionnaire submits questionnaire response data for authenticated anonymous users
func (h *Handler) SubmitQuestionnaire(ctx echo.Context) error {
	req, err := BindAndValidate[models.SubmitQuestionnaireRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	var (
		assessmentID string
		email        string
		internalCtx  context.Context
		ownerID      string
		isAnonymous  bool
	)

	internalCtx = reqCtx

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok {
		logx.FromContext(reqCtx).Error().Msg("error getting authenticated user")
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if anonAssessmentID, ok := auth.ActiveAssessmentIDKey.Get(reqCtx); ok {
		assessmentID = anonAssessmentID

		email = caller.SubjectEmail
		ownerID = caller.OrganizationID
		internalCtx = auth.WithInternalOperationContext(auth.WithCaller(internalCtx, caller))

		if email == "" {
			isAnonymous = true
		}

		internalCtx = auth.ActiveAssessmentIDKey.Set(internalCtx, assessmentID)
	} else {

		// for regular/normal authenticated users, we expect the assessment id to be passed
		// in the request by the client.
		//
		// for anon users, it's embedded inside the jwt already
		if req.AssessmentID == "" {
			return h.BadRequest(ctx, ErrMissingAssessmentID)
		}

		assessmentID = req.AssessmentID
		email = caller.SubjectEmail
		ownerID = caller.OrganizationID

		// bypass FGA tuple creation for questionnaire submissions;
		// DocumentData ownership is tracked via AssessmentResponse, not FGA tuples
		internalCtx = auth.WithCaller(internalCtx, &auth.Caller{
			OrganizationID: caller.OrganizationID,
			SubjectID:      caller.SubjectID,
			Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
		})
	}

	if assessmentID == "" {
		return h.BadRequest(ctx, ErrMissingAssessmentID)
	}

	if len(req.Data) == 0 {
		return h.BadRequest(ctx, ErrMissingQuestionnaireData)
	}

	if isAnonymous && req.IsDraft {
		return h.BadRequest(ctx, ErrAnonymousQuestionnaireDraft)
	}

	assessment, err := h.DBClient.Assessment.Query().
		Where(assessment.IDEQ(assessmentID)).
		WithAssessmentResponses().
		Only(internalCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentNotFound)
		}

		logx.FromContext(reqCtx).Err(err).Msg("could not fetch assessment")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if ownerID == "" {
		ownerID = assessment.OwnerID
	}

	var assessmentResponse *generated.AssessmentResponse

	// preview tokens resolve to the test response; real recipients get their real response
	isPreview, _ := auth.ActiveAssessmentPreviewKey.Get(reqCtx)

	if email != "" {
		query := h.DBClient.AssessmentResponse.Query().
			Where(assessmentresponse.EmailEqualFold(email),
				assessmentresponse.AssessmentIDEQ(assessmentID),
				assessmentresponse.IsTestEQ(isPreview))

		var campaignPredicate = assessmentresponse.CampaignIDIsNil()

		if id, ok := auth.ActiveCampaignIDKey.Get(reqCtx); ok && id != "" {
			campaignPredicate = assessmentresponse.CampaignIDEQ(id)
		}

		assessmentResponse, err = query.Where(campaignPredicate).Only(internalCtx)
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, ErrAssessmentResponseNotFound)
		}

		if err != nil {
			return h.NotFound(ctx, err)
		}
	}

	if assessmentResponse != nil && assessmentResponse.Status == enums.AssessmentResponseStatusCompleted {
		return h.BadRequest(ctx, ErrAssessmentResponseAlreadyCompleted)
	}

	var documentDataID string

	if assessmentResponse != nil && assessmentResponse.DocumentDataID != "" {
		err = h.DBClient.DocumentData.UpdateOneID(assessmentResponse.DocumentDataID).
			SetData(req.Data).
			Exec(internalCtx)
		if err != nil {
			logx.FromContext(reqCtx).Err(err).Msg("could not update document data")
			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		documentDataID = assessmentResponse.DocumentDataID
	} else {
		documentDataQuery := h.DBClient.DocumentData.Create().
			SetOwnerID(ownerID)

		if assessment.TemplateID != "" {
			documentDataQuery = documentDataQuery.SetTemplateID(assessment.TemplateID)
		}

		documentData, err := documentDataQuery.SetData(req.Data).Save(internalCtx)
		if err != nil {
			logx.FromContext(reqCtx).Err(err).Msg("could not create document data")
			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		documentDataID = documentData.ID
	}

	if assessmentResponse == nil {
		assessmentResponse, err = h.DBClient.AssessmentResponse.Create().
			SetAssessmentID(assessmentID).
			SetOwnerID(ownerID).
			Save(internalCtx)
		if err != nil {
			logx.FromContext(reqCtx).Err(err).Msg("could not create assessment response")
			return h.InternalServerError(ctx, ErrProcessingRequest)
		}
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

	freshResponse, err := responseUpdate.Save(internalCtx)
	if err != nil {
		if errors.Is(err, hooks.ErrAssessmentInCompleted) {
			return h.BadRequest(ctx, err)
		}

		logx.FromContext(reqCtx).Err(err).Msg("could not update assessment response")
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	response := models.SubmitQuestionnaireResponse{
		DocumentDataID: documentDataID,
		Status:         freshResponse.Status.String(),
	}

	if !freshResponse.CompletedAt.IsZero() {
		response.CompletedAt = freshResponse.CompletedAt.Format(time.RFC3339)
	}

	return h.Success(ctx, response)
}
