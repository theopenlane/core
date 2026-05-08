package handlers

import (
	"net/url"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/urlx"
)

const maxQuestionnaireResendAttempts = 5

// ResendQuestionnaireEmail handles requests to resend the questionnaire authentication email with a new JWT token
func (h *Handler) ResendQuestionnaireEmail(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleResendQuestionnaireRequest, models.ExampleResendQuestionnaireResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, auth.NewWebhookCaller(""))

	out := &models.ResendReply{
		Reply:   rout.Reply{Success: true},
		Message: "If the email address is associated with an active assessment, a new link has been sent to access this",
	}

	assessmentResp, err := h.DBClient.AssessmentResponse.Query().
		Where(
			assessmentresponse.EmailEqualFold(in.Email),
			assessmentresponse.AssessmentIDEQ(in.AssessmentID),
		).
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.Success(ctx, out, openapi)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error querying assessment response")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if assessmentResp.Status == enums.AssessmentResponseStatusCompleted {
		return h.Success(ctx, out, openapi)
	}

	if !assessmentResp.DueDate.IsZero() && time.Now().After(assessmentResp.DueDate) {
		return h.BadRequest(ctx, ErrAssessmentResponseOverdue, openapi)
	}

	if assessmentResp.SendAttempts >= maxQuestionnaireResendAttempts {
		return h.Success(ctx, out, openapi)
	}

	assessmentResp, err = h.DBClient.AssessmentResponse.UpdateOneID(assessmentResp.ID).
		SetSendAttempts(assessmentResp.SendAttempts + 1).
		Save(allowCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error incrementing send attempts")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	assessmentData, err := h.DBClient.Assessment.Query().
		Where(assessment.ID(in.AssessmentID)).
		Select(assessment.FieldName).
		Only(allowCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error querying assessment")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	baseURL, err := url.Parse(h.IntegrationsConfig.Email.ProductURL + "/questionnaire")
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error parsing questionnaire base URL")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	duration := h.DBClient.TokenManager.Config().AssessmentAccessDuration

	result, err := urlx.GenerateAnonTokenURL(reqCtx, h.DBClient.TokenManager, h.DBClient.Shortlinks, *baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonQuestionnaireJWTPrefix,
		SubjectID: ulids.New().String(),
		OrgID:     assessmentResp.OwnerID,
		Email:     in.Email,
		Duration:  duration,
		ExtraClaims: func(c *tokens.Claims) {
			c.AssessmentID = in.AssessmentID
		},
	})
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error generating questionnaire auth URL")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if err := h.sendEmail(reqCtx, email.QuestionnaireAuthOp.Name(), email.QuestionnaireAuthEmail{
		RecipientInfo:  email.RecipientInfo{Email: in.Email},
		AssessmentName: assessmentData.Name,
		AuthURL:        result.URL,
	}); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error sending questionnaire auth email")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, out, openapi)
}
