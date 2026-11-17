package email

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/newman"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessment"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
	"github.com/theopenlane/core/v2/internal/integrations/templatekit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/shortlinks"
	"github.com/theopenlane/core/v2/pkg/urlx"
	"github.com/theopenlane/logx"
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
	camp       *generated.Campaign
	dispatcher Dispatcher
	// defaults are the linked template's authored values, nil when the campaign uses the default wording
	defaults map[string]any
	// overlay is the campaign template context
	overlay CampaignContext
	// base carries the org and questionnaire names shared by every recipient
	base QuestionnaireAuthEmail
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
		camp:       camp,
		dispatcher: questionnaireAccess,
		overlay:    campaignOverlay(camp),
		base: QuestionnaireAuthEmail{
			OrgName:        organizationDisplayName(ctx, db, camp.OwnerID),
			AssessmentName: questionnaireName,
		},
	}

	template := camp.Edges.EmailTemplate
	if template == nil {
		return dispatch, nil
	}

	dispatcher, err := campaignDispatcher(camp)
	if err != nil {
		return dispatch, err
	}

	dispatch.dispatcher = dispatcher
	dispatch.defaults = template.Defaults

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

	// set the base settings and then the defaults
	input := dispatch.base
	input.Email = target.Email
	input.FirstName, input.LastName = splitFullName(target.FullName)
	input.Tags = tags
	input.AuthURL = authURL

	payload, err := questionnairePayload(dispatch, input)
	if err != nil {
		return err
	}

	if err := dispatch.dispatcher.SendByKey(ctx, req, client, payload); err != nil {
		return err
	}

	if target.ID == "" {
		return nil
	}

	return markCampaignTargetSent(ctx, req.DB, target.ID)
}

// questionnairePayload merges the template defaults with the per-send fields
// system keys are dropped from the defaults so a template cannot change the recipient, questionnaire, or access link
func questionnairePayload(dispatch questionnaireDispatch, input QuestionnaireAuthEmail) (json.RawMessage, error) {
	defaults := lo.OmitByKeys(dispatch.defaults, slices.Concat(baseSystemKeys, questionnaireAccess.SystemKeys))

	return templatekit.BuildDispatchPayload(defaults, dispatch.overlay, input)
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
