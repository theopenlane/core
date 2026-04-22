package email

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/urlx"
	"github.com/theopenlane/iam/tokens"
)

// SendQuestionnaireCampaignRequest is the operation config for dispatching a questionnaire campaign
type SendQuestionnaireCampaignRequest struct {
	// CampaignID is the identifier of the campaign to dispatch
	CampaignID string `json:"campaignId" jsonschema:"required,description=Campaign identifier"`
}

// SendQuestionnaireCampaign dispatches questionnaire access emails to all pending campaign targets,
// creating assessment responses and generating anonymous access tokens for each recipient
type SendQuestionnaireCampaign struct{}

// Handle returns the typed operation handler for builder registration
func (s SendQuestionnaireCampaign) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, SendQuestionnaireCampaignOp, ErrCampaignNotFound, s.Run)
}

// Run loads the campaign and its assessment, iterates pending targets, creates assessment
// responses, generates anonymous JWT access tokens, and sends questionnaire access emails.
// Failed targets are logged and processing continues so a single failure does not abort the dispatch
func (SendQuestionnaireCampaign) Run(ctx context.Context, req types.OperationRequest, client *EmailClient, cfg SendQuestionnaireCampaignRequest) (json.RawMessage, error) {
	camp, err := req.DB.Campaign.Query().
		Where(campaign.IDEQ(cfg.CampaignID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, ErrCampaignNotFound
		}

		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", cfg.CampaignID).Msg("failed loading campaign for questionnaire dispatch")

		return nil, ErrCampaignNotFound
	}

	if camp.AssessmentID == "" {
		return nil, ErrCampaignMissingAssessment
	}

	assessmentObj, err := req.DB.Assessment.Query().
		Where(assessment.IDEQ(camp.AssessmentID)).
		Select(assessment.FieldName).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("assessment_id", camp.AssessmentID).Msg("failed loading assessment for questionnaire dispatch")

		return nil, fmt.Errorf("%w: %w", ErrAssessmentNotFound, err)
	}

	targets, err := req.DB.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(cfg.CampaignID),
			campaigntarget.SentAtIsNil(),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", cfg.CampaignID).Msg("failed loading campaign targets")

		return nil, err
	}

	for _, target := range targets {
		if err := sendQuestionnaireToTarget(ctx, req, req.DB, client, camp, assessmentObj.Name, target); err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("campaign_id", cfg.CampaignID).
				Str("target_id", target.ID).
				Msg("failed dispatching questionnaire email to target")
		}
	}

	return nil, nil
}

// sendQuestionnaireToTarget creates an assessment response, generates an anonymous access
// token URL, dispatches the questionnaire access email through the questionnaireAuthEmail
// operation, and marks the target as sent
func sendQuestionnaireToTarget(ctx context.Context, req types.OperationRequest, db *generated.Client, client *EmailClient, camp *generated.Campaign, assessmentName string, target *generated.CampaignTarget) error {
	create := db.AssessmentResponse.Create().
		SetAssessmentID(camp.AssessmentID).
		SetCampaignID(camp.ID).
		SetEmail(target.Email).
		SetOwnerID(camp.OwnerID)

	if camp.EntityID != "" {
		create.SetEntityID(camp.EntityID)
	}

	if camp.DueDate != nil && !camp.DueDate.IsZero() {
		create.SetDueDate(time.Time(*camp.DueDate))
	}

	if err := create.Exec(ctx); err != nil {
		return fmt.Errorf("create assessment response: %w", err)
	}

	baseURL, err := url.Parse(client.Config.ProductURL + "/questionnaire")
	if err != nil {
		return fmt.Errorf("parse questionnaire URL: %w", err)
	}

	anonSubjectID := ulids.New().String()
	duration := db.TokenManager.Config().AssessmentAccessDuration

	result, err := urlx.GenerateAnonTokenURL(ctx, db.TokenManager, db.Shortlinks, *baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonQuestionnaireJWTPrefix,
		SubjectID: anonSubjectID,
		OrgID:     camp.OwnerID,
		Email:     target.Email,
		Duration:  duration,
		ExtraClaims: func(c *tokens.Claims) {
			c.AssessmentID = camp.AssessmentID
		},
	})
	if err != nil {
		return fmt.Errorf("generate questionnaire token URL: %w", err)
	}

	input := QuestionnaireAuthEmail{
		RecipientInfo: RecipientInfo{
			Email: target.Email,
			Tags:  []newman.Tag{{Name: TagCampaignTargetID, Value: target.ID}},
		},
		AssessmentName: assessmentName,
		AuthURL:        result.URL,
	}

	if err := questionnaireAuthEmail.dispatch(ctx, req, client, input); err != nil {
		return err
	}

	now := models.DateTime(time.Now())
	if err := db.CampaignTarget.UpdateOneID(target.ID).SetSentAt(now).Exec(ctx); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}

	return nil
}
