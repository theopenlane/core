package email

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/urlx"
	"github.com/theopenlane/iam/tokens"
)

// SendQuestionnaireCampaignRequest is the operation config for dispatching a questionnaire campaign
type SendQuestionnaireCampaignRequest struct {
	CampaignDispatchInput
	// TestEmail dispatches a single test questionnaire email without adding a campaign target
	TestEmail string `json:"testEmail,omitempty" jsonschema:"description=Recipient email for a test questionnaire send"`
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
// Returns a marshaled CampaignDispatchResult with counts
func (SendQuestionnaireCampaign) Run(ctx context.Context, req types.OperationRequest, client *EmailClient, cfg SendQuestionnaireCampaignRequest) (json.RawMessage, error) {
	camp, dispatchable, skipped, err := loadCampaignWithTargets(ctx, req.DB, cfg.CampaignDispatchInput)
	if err != nil {
		return nil, err
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

	if cfg.TestEmail != "" {
		if err := sendQuestionnaireToRecipient(ctx, req, req.DB, client, camp, assessmentObj.Name, cfg.TestEmail, "", true); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed dispatching questionnaire test email")

			return nil, fmt.Errorf("%w: %s", ErrQuestionnaireDispatchFailed, cfg.TestEmail)
		}

		return json.Marshal(CampaignDispatchResult{SentCount: 1})
	}

	return dispatchCampaignTargets(ctx, dispatchable, skipped, func(ctx context.Context, target *generated.CampaignTarget) error {
		return sendQuestionnaireToRecipient(ctx, req, req.DB, client, camp, assessmentObj.Name, target.Email, target.ID, false)
	})
}

// sendQuestionnaireToRecipient creates an assessment response, generates an anonymous access
// token URL, dispatches the questionnaire access email through the questionnaireAuthEmail
// operation, and marks campaign targets as sent for non-test dispatches
func sendQuestionnaireToRecipient(ctx context.Context, req types.OperationRequest, db *generated.Client, client *EmailClient, camp *generated.Campaign, assessmentName string, email string, campaignTargetID string, isTest bool) error {
	response, err := createAssessmentResponseForRecipient(ctx, db, camp, camp.AssessmentID, email, isTest)
	if err != nil {
		return err
	}

	authURL, err := questionnaireAuthURL(ctx, db, client, camp, email)
	if err != nil {
		return err
	}

	tags := make([]newman.Tag, 0, 3)
	tags = append(tags, newman.Tag{Name: TagAssessmentResponseID, Value: response.ID})

	if campaignTargetID != "" {
		tags = append(tags, newman.Tag{Name: TagCampaignTargetID, Value: campaignTargetID})
	}

	if isTest {
		tags = append(tags, newman.Tag{Name: TagIsTest, Value: "true"})
	}

	input := QuestionnaireAuthEmail{
		RecipientInfo: RecipientInfo{
			Email: email,
			Tags:  tags,
		},
		AssessmentName: assessmentName,
		AuthURL:        authURL,
	}

	if err := questionnaireAuthEmail.dispatch(ctx, req, client, input); err != nil {
		return err
	}

	if campaignTargetID == "" {
		return nil
	}

	return markCampaignTargetSent(ctx, db, campaignTargetID)
}

// questionnaireAuthURL generates an anonymous access token URL for the campaign's assessment questionnaire
func questionnaireAuthURL(ctx context.Context, db *generated.Client, client *EmailClient, camp *generated.Campaign, recipientEmail string) (string, error) {
	baseURL, err := url.Parse(client.Config.ProductURL + "/questionnaire")
	if err != nil {
		return "", fmt.Errorf("parse questionnaire URL: %w", err)
	}

	result, err := urlx.GenerateAnonTokenURL(ctx, db.TokenManager, db.Shortlinks, *baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonQuestionnaireJWTPrefix,
		SubjectID: ulids.New().String(),
		OrgID:     camp.OwnerID,
		Email:     recipientEmail,
		Duration:  db.TokenManager.Config().AssessmentAccessDuration,
		ExtraClaims: func(c *tokens.Claims) {
			c.AssessmentID = camp.AssessmentID
		},
	})
	if err != nil {
		return "", fmt.Errorf("generate questionnaire token URL: %w", err)
	}

	return result.URL, nil
}
