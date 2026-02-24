package handlers

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/logx"
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

	org, err := h.DBClient.Organization.Query().
		Where(organization.ID(assessmentResp.OwnerID)).
		Select(organization.FieldDisplayName).
		Only(allowCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error querying organization")

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

	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonQuestionnaireJWTPrefix, ulids.New().String())

	newClaims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:       anonUserID,
		OrgID:        assessmentResp.OwnerID,
		AssessmentID: in.AssessmentID,
		Email:        in.Email,
	}

	duration := h.DBClient.TokenManager.Config().AssessmentAccessDuration

	accessToken, _, err := h.DBClient.TokenManager.CreateTokenPair(newClaims, tokens.WithAccessDuration(duration))
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error creating token pair")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	authURL, err := h.shortenQuestionnaireURL(reqCtx, accessToken)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error shortening questionnaire URL")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	email, err := h.Emailer.NewQuestionnaireAuthEmail(emailtemplates.Recipient{
		Email: in.Email,
	}, accessToken, emailtemplates.QuestionnaireAuthData{
		CompanyName:              org.DisplayName,
		AssessmentName:           assessmentData.Name,
		QuestionnaireAuthFullURL: authURL,
	})
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error creating questionnaire auth email")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	tags := []newman.Tag{
		{Name: "assessment_response_id", Value: assessmentResp.ID},
	}

	if assessmentResp.CampaignID != "" {
		tags = append(tags, newman.Tag{Name: "campaign_id", Value: assessmentResp.CampaignID})
	}

	email.Tags = append(email.Tags, tags...)

	if _, err = h.DBClient.Job.Insert(reqCtx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error queuing questionnaire auth email")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, out, openapi)
}

func (h *Handler) shortenQuestionnaireURL(ctx context.Context, token string) (string, error) {
	baseURL, err := url.Parse(h.Emailer.URLS.Questionnaire)
	if err != nil {
		return "", err
	}

	originalURL := baseURL.ResolveReference(&url.URL{RawQuery: url.Values{"token": []string{token}}.Encode()})

	if h.ShortlinksClient == nil {
		return originalURL.String(), nil
	}

	shortURL, err := h.ShortlinksClient.Create(ctx, originalURL.String(), "")
	if err != nil {
		logx.FromContext(ctx).Error().Str("baseURL", baseURL.String()).
			Err(err).
			Msg("failed to shorten URL, using original")

		return originalURL.String(), nil
	}

	return shortURL, nil
}
