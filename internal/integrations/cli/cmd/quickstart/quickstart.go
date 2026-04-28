package quickstart

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/logx"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/go-client/graphclient"
)

// quickstartTag marks resources created by this command so they are easy to
// spot (and clean up) later
const quickstartTag = "cli-quickstart"

// recipientFullName is the test recipient name used for all quickstart targets
const recipientFullName = "Flynne Fisher"

// defaultsJSON is the canonical BrandedMessageRequest payload used to seed the
// quickstart template. Edit defaults.example.json to change the wording
//
//go:embed defaults.example.json
var defaultsJSON []byte

var command = &cobra.Command{
	Use:   "quickstart",
	Short: "end-to-end Openlane Campaigns smoke test: branded + questionnaire campaigns",
	Long: `Provisions two Openlane Campaigns pipelines end-to-end:

1. Branded campaign: creates a template from defaults.example.json, a CUSTOM
   campaign targeting the recipient, and launches it through the email provider.

2. Questionnaire campaign: creates an assessment, a QUESTIONNAIRE campaign
   linked to the assessment, and launches the questionnaire dispatch flow.

All resources are tagged "cli-quickstart" for easy discovery and cleanup.

Override --to to send to a different address.`,
	RunE: func(c *cobra.Command, _ []string) error {
		return run(c.Context())
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().String("to", "", "recipient email (defaults to openlane.auth.email)")
}

// run executes the end-to-end flow for both campaign types
func run(ctx context.Context) error {
	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	recipient, err := resolveRecipient()
	if err != nil {
		return err
	}

	brandedResult, err := runBrandedCampaign(ctx, client, recipient)
	if err != nil {
		return err
	}

	questionnaireResult, err := runQuestionnaireCampaign(ctx, client, recipient)
	if err != nil {
		return err
	}

	headers := []string{"Type", "CampaignID", "Recipient", "Queued", "Skipped"}
	rows := [][]string{
		{
			"branded",
			brandedResult.campaignID,
			recipient,
			strconv.FormatInt(brandedResult.queued, 10),
			strconv.FormatInt(brandedResult.skipped, 10),
		},
		{
			"questionnaire",
			questionnaireResult.campaignID,
			recipient,
			strconv.FormatInt(questionnaireResult.queued, 10),
			strconv.FormatInt(questionnaireResult.skipped, 10),
		},
	}

	return cmd.RenderTable(nil, headers, rows)
}

// launchResult holds the output of a campaign launch
type launchResult struct {
	campaignID string
	queued     int64
	skipped    int64
}

// runBrandedCampaign creates a template, campaign, and launches the branded dispatch
func runBrandedCampaign(ctx context.Context, client *openlaneclient.Client, recipient string) (*launchResult, error) {
	templateKey := email.BrandedMessageOp.Name()

	templateID, err := createBrandedTemplate(ctx, client, templateKey)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create branded email template")
		return nil, fmt.Errorf("create branded template: %w", err)
	}

	campaignID, err := createBrandedCampaign(ctx, client, templateID, recipient)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create branded campaign")
		return nil, fmt.Errorf("create branded campaign: %w", err)
	}

	launch, err := client.LaunchCampaign(ctx, graphclient.LaunchCampaignInput{CampaignID: campaignID})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to launch branded campaign")
		return nil, fmt.Errorf("launch branded campaign: %w", err)
	}

	logx.FromContext(ctx).Info().Str("campaign_id", campaignID).Msg("branded campaign launched")

	return &launchResult{
		campaignID: campaignID,
		queued:     launch.LaunchCampaign.QueuedCount,
		skipped:    launch.LaunchCampaign.SkippedCount,
	}, nil
}

// runQuestionnaireCampaign creates an assessment, campaign, and launches the questionnaire dispatch
func runQuestionnaireCampaign(ctx context.Context, client *openlaneclient.Client, recipient string) (*launchResult, error) {
	assessmentID, err := createAssessment(ctx, client)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create assessment")
		return nil, fmt.Errorf("create assessment: %w", err)
	}

	campaignID, err := createQuestionnaireCampaign(ctx, client, assessmentID, recipient)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create questionnaire campaign")
		return nil, fmt.Errorf("create questionnaire campaign: %w", err)
	}

	launch, err := client.LaunchCampaign(ctx, graphclient.LaunchCampaignInput{CampaignID: campaignID})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to launch questionnaire campaign")
		return nil, fmt.Errorf("launch questionnaire campaign: %w", err)
	}

	logx.FromContext(ctx).Info().Str("campaign_id", campaignID).Msg("questionnaire campaign launched")

	return &launchResult{
		campaignID: campaignID,
		queued:     launch.LaunchCampaign.QueuedCount,
		skipped:    launch.LaunchCampaign.SkippedCount,
	}, nil
}

// resolveRecipient picks the recipient from --to or the authed user's email
func resolveRecipient() (string, error) {
	if to := cmd.Config.String("to"); to != "" {
		return to, nil
	}

	if email := cmd.Config.String("openlane.auth.email"); email != "" {
		return email, nil
	}

	return "", ErrRecipientResolution
}

// loadDefaults unmarshals the embedded defaults.example.json into the map shape
// expected by graphclient.CreateEmailTemplateInput.Defaults
func loadDefaults() (map[string]any, error) {
	out := map[string]any{}
	if err := json.Unmarshal(defaultsJSON, &out); err != nil {
		return nil, fmt.Errorf("parse quickstart defaults: %w", err)
	}

	return out, nil
}

// createBrandedTemplate creates a CAMPAIGN_RECIPIENT template bound to the branded-message catalog entry
func createBrandedTemplate(ctx context.Context, client *openlaneclient.Client, key string) (string, error) {
	defaults, err := loadDefaults()
	if err != nil {
		return "", err
	}

	input := graphclient.CreateEmailTemplateInput{
		Key:             key,
		Name:            "Quickstart branded template",
		TemplateContext: lo.ToPtr(enums.TemplateContextCampaignRecipient),
		Defaults:        defaults,
		Active:          lo.ToPtr(true),
	}

	resp, err := client.CreateEmailTemplate(ctx, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", key).Msg("failed to create email template")
		return "", err
	}

	id := resp.CreateEmailTemplate.EmailTemplate.ID
	logx.FromContext(ctx).Debug().Str("template_id", id).Str("template_key", key).Msg("created branded email template")

	return id, nil
}

// createBrandedCampaign creates a POLICY_ATTESTATION campaign targeting the recipient with an email template
func createBrandedCampaign(ctx context.Context, client *openlaneclient.Client, templateID, recipient string) (string, error) {
	dueDate := models.DateTime(time.Now().Add(7 * 24 * time.Hour))

	campaign := &graphclient.CreateCampaignInput{
		Name:            "Openlane Campaigns",
		Description:     lo.ToPtr("Branded campaign end-to-end smoke test"),
		CampaignType:    lo.ToPtr(enums.CampaignTypePolicyAttestation),
		EmailTemplateID: &templateID,
		DueDate:         &dueDate,
		Tags:            []string{quickstartTag},
	}

	input := graphclient.CreateCampaignWithTargetsInput{
		Campaign: campaign,
		Targets: []*graphclient.CreateCampaignTargetInput{{
			Email:    recipient,
			FullName: lo.ToPtr(recipientFullName),
		}},
	}

	resp, err := client.CreateCampaignWithTargets(ctx, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("recipient", recipient).Msg("failed to create branded campaign")
		return "", err
	}

	id := resp.CreateCampaignWithTargets.Campaign.ID
	logx.FromContext(ctx).Debug().Str("campaign_id", id).Str("recipient", recipient).Msg("created branded campaign")

	return id, nil
}

// createAssessment creates a questionnaire assessment for the quickstart
func createAssessment(ctx context.Context, client *openlaneclient.Client) (string, error) {
	input := graphclient.CreateAssessmentInput{
		Name:           "Quickstart assessment",
		AssessmentType: lo.ToPtr(enums.AssessmentTypeInternal),
		Tags:           []string{quickstartTag},
		Jsonconfig: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"acknowledged": map[string]any{
					"type":        "boolean",
					"title":       "Acknowledgement",
					"description": "I confirm I have reviewed and understood this questionnaire",
				},
			},
			"required": []string{"acknowledged"},
		},
	}

	resp, err := client.CreateAssessment(ctx, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create assessment")
		return "", err
	}

	id := resp.CreateAssessment.Assessment.ID
	logx.FromContext(ctx).Debug().Str("assessment_id", id).Msg("created assessment")

	return id, nil
}

// createQuestionnaireCampaign creates a QUESTIONNAIRE campaign linked to the assessment
func createQuestionnaireCampaign(ctx context.Context, client *openlaneclient.Client, assessmentID, recipient string) (string, error) {
	dueDate := models.DateTime(time.Now().Add(7 * 24 * time.Hour))

	campaign := &graphclient.CreateCampaignInput{
		Name:         "Acknowledge Matt is the Greatest",
		Description:  lo.ToPtr("Questionnaire campaign end-to-end smoke test"),
		CampaignType: lo.ToPtr(enums.CampaignTypeQuestionnaire),
		AssessmentID: &assessmentID,
		DueDate:      &dueDate,
		Tags:         []string{quickstartTag},
	}

	input := graphclient.CreateCampaignWithTargetsInput{
		Campaign: campaign,
		Targets: []*graphclient.CreateCampaignTargetInput{{
			Email:    recipient,
			FullName: lo.ToPtr(recipientFullName),
		}},
	}

	resp, err := client.CreateCampaignWithTargets(ctx, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("recipient", recipient).Msg("failed to create questionnaire campaign")
		return "", err
	}

	id := resp.CreateCampaignWithTargets.Campaign.ID
	logx.FromContext(ctx).Debug().Str("campaign_id", id).Str("assessment_id", assessmentID).Msg("created questionnaire campaign")

	return id, nil
}
