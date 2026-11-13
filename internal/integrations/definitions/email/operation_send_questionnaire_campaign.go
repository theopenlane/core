package email

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessment"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/shortlinks"
	"github.com/theopenlane/core/v2/pkg/urlx"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/jsonx"
)

// maxQuestionnaireTags is the pre-allocation capacity for per-message newman tags
const maxQuestionnaireTags = 3

// SendQuestionnaireCampaignRequest is the operation config for dispatching a questionnaire campaign
type SendQuestionnaireCampaignRequest struct {
	CampaignDispatchInput
	// TestEmail dispatches a single test questionnaire email without adding a campaign target
	TestEmail string `json:"testEmail,omitempty" jsonschema:"description=Recipient email for a test questionnaire send"`
}

// SendQuestionnaireCampaign dispatches questionnaire access emails to all pending campaign targets, creating assessment responses and generating anonymous access tokens for each recipient
type SendQuestionnaireCampaign struct{}

// SendQuestionnaireCampaignOp is the operation ref for the questionnaire campaign dispatch operation
var SendQuestionnaireCampaignOp = types.OperationPayloadOf[SendQuestionnaireCampaignRequest]().Handles(SendQuestionnaireCampaign{}.Run).Policy(types.ExecutionPolicy{SkipRunRecord: true})

// questionnaireDispatch holds the campaign-wide state shared by every questionnaire send
type questionnaireDispatch struct {
	camp              *generated.Campaign
	questionnaireName string
	// orgName is the sending organization's name
	orgName string
	// content is the linked template's authored text and branding, empty when the campaign uses the default wording
	content QuestionnaireEmailContent
	// overlay is the campaign template context
	overlay CampaignContext
}

// questionnaireCampaignEdges eager-loads the email template linked to a questionnaire campaign
func questionnaireCampaignEdges(q *generated.CampaignQuery) {
	q.WithEmailTemplate()
}

// Run loads the campaign and its assessment, iterates pending targets, creates assessment responses, generates anonymous JWT access tokens, and sends questionnaire access emails
func (SendQuestionnaireCampaign) Run(ctx context.Context, req types.OperationRequest, client *Client, cfg SendQuestionnaireCampaignRequest) (json.RawMessage, error) {
	camp, dispatchable, skipped, err := loadCampaignWithTargets(ctx, req.DB, cfg.CampaignDispatchInput, questionnaireCampaignEdges)
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

	dispatch, err := newQuestionnaireDispatch(ctx, req.DB, camp, assessmentObj.Name)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", camp.ID).Msg("failed resolving questionnaire email template")

		return nil, err
	}

	if cfg.TestEmail != "" {
		if err := sendQuestionnaireToRecipient(ctx, req, client, dispatch, &generated.CampaignTarget{Email: cfg.TestEmail}, true); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed dispatching questionnaire test email")

			return nil, fmt.Errorf("%w: %s", ErrQuestionnaireDispatchFailed, cfg.TestEmail)
		}

		return json.Marshal(CampaignDispatchResult{SentCount: 1})
	}

	return dispatchCampaignTargets(ctx, dispatchable, skipped, func(ctx context.Context, target *generated.CampaignTarget) error {
		return sendQuestionnaireToRecipient(ctx, req, client, dispatch, target, false)
	})
}

// newQuestionnaireDispatch resolves the campaign-wide send state, leaving content empty when no template is linked
func newQuestionnaireDispatch(ctx context.Context, db *generated.Client, camp *generated.Campaign, questionnaireName string) (questionnaireDispatch, error) {
	dispatch := questionnaireDispatch{
		camp:              camp,
		questionnaireName: questionnaireName,
		orgName:           organizationDisplayName(ctx, db, camp.OwnerID),
		overlay:           campaignOverlay(camp),
	}

	template := camp.Edges.EmailTemplate
	if template == nil {
		return dispatch, nil
	}

	if err := ValidateCampaignTemplateKey(camp.CampaignType, template.Key); err != nil {
		return dispatch, err
	}

	if err := jsonx.RoundTrip(template.Defaults, &dispatch.content); err != nil {
		return dispatch, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	return dispatch, nil
}

// sendQuestionnaireToRecipient creates an assessment response, generates an anonymous access token URL, dispatches the questionnaire access email, and marks campaign targets as sent for non-test dispatches
func sendQuestionnaireToRecipient(ctx context.Context, req types.OperationRequest, client *Client, dispatch questionnaireDispatch, target *generated.CampaignTarget, isTest bool) error {
	if strings.TrimSpace(target.Email) == "" {
		return nil
	}

	camp := dispatch.camp

	response, err := createAssessmentResponseForRecipient(ctx, req.DB, camp, camp.AssessmentID, target.Email, isTest)
	if err != nil {
		return err
	}

	authURL, err := QuestionnaireAuthURL(ctx, req.DB, camp.AssessmentID, camp.OwnerID, target.Email, camp.ID, isTest)
	if err != nil {
		return err
	}

	tags := make([]newman.Tag, 0, maxQuestionnaireTags)
	tags = append(tags, newman.Tag{Name: TagAssessmentResponseID, Value: response.ID})

	if target.ID != "" {
		tags = append(tags, newman.Tag{Name: TagCampaignTargetID, Value: target.ID})
	}

	if isTest {
		tags = append(tags, newman.Tag{Name: TagIsTest, Value: TagTrueValue})
	}

	input := QuestionnaireAuthEmail{
		Email:          target.Email,
		Tags:           tags,
		AssessmentName: dispatch.questionnaireName,
		AuthURL:        authURL,
	}

	if err := dispatchQuestionnaireEmail(ctx, req, client, dispatch, target, input); err != nil {
		return err
	}

	if target.ID == "" {
		return nil
	}

	return markCampaignTargetSent(ctx, req.DB, target.ID)
}

// dispatchQuestionnaireEmail sends the questionnaire invitation, using the linked template's content when the campaign has one
// the template only supplies the authored content, so it cannot change the recipient, questionnaire, or access link
func dispatchQuestionnaireEmail(ctx context.Context, req types.OperationRequest, client *Client, dispatch questionnaireDispatch, target *generated.CampaignTarget, input QuestionnaireAuthEmail) error {
	input.FirstName, input.LastName = splitFullName(target.FullName)
	input.OrgName = dispatch.orgName

	payload, err := json.Marshal(QuestionnaireInvitationRequest{
		QuestionnaireAuthEmail:    input,
		CampaignContext:           dispatch.overlay,
		QuestionnaireEmailContent: dispatch.content,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	return questionnaireInvitation.SendByKey(ctx, req, client, payload)
}

// questionnairePath is the product route serving questionnaire access links
const questionnairePath = "questionnaire"

// QuestionnaireAuthURL generates an anonymous access token URL for questionnaire access
func QuestionnaireAuthURL(ctx context.Context, db *generated.Client, assessmentID, ownerID, recipientEmail, campaignID string, isTest bool) (string, error) {
	productURL, err := urlx.ParseAbsolute(db.EntConfig.QuestionnaireProductURL)
	if err != nil {
		return "", fmt.Errorf("parse questionnaire URL: %w", err)
	}

	baseURL := productURL.JoinPath(questionnairePath)

	result, err := urlx.GenerateAnonTokenURL(ctx, db.TokenManager, db.Shortlinks, *baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonQuestionnaireJWTPrefix,
		SubjectID: assessmentID,
		OrgID:     ownerID,
		Email:     recipientEmail,
		Purpose:   shortlinks.PurposeQuestionnaireAuth,
		Duration:  db.TokenManager.Config().AssessmentAccessDuration,
		ExtraClaims: func(c *tokens.Claims) {
			c.AssessmentID = assessmentID
			c.AssessmentPreview = isTest

			if campaignID != "" {
				c.CampaignID = campaignID
			}
		},
	})
	if err != nil {
		return "", fmt.Errorf("generate questionnaire token URL: %w", err)
	}

	return result.URL, nil
}
