package quickstart

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/core/pkg/logx"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/go-client/graphclient"
)

// quickstartTag marks resources created by this command so they are easy to
// spot (and clean up) later
const quickstartTag = "cli-quickstart"

// resourcePrefix is prepended to generated resource names
const resourcePrefix = "cli-quickstart-"

// defaultSubject is the subject line for the smoke-test template
const defaultSubject = "Openlane CLI quickstart test"

// defaultBody is the body template for the smoke-test email. Flat keys match
// the variables produced by sendCampaignTargetEmail + buildTemplateData
// (recipientFirstName, campaignName, companyName, ...); the body contains
// content blocks only — the theme renderer supplies the surrounding HTML
const defaultBody = `Hi {{ .recipientFirstName }},

This is a smoke-test email from the Openlane integrations CLI for campaign **{{ .campaignName }}**.

If you received this, the end-to-end flow (branding + template + campaign + launch + dispatch) is working.
`

var command = &cobra.Command{
	Use:   "quickstart",
	Short: "end-to-end smoke test: create branding + template + campaign, then launch",
	Long: `Creates a timestamped email branding, email template, and campaign
targeted at the authenticated user, then launches the campaign. All resources
are tagged "cli-quickstart" for easy discovery and cleanup.

Override --to to send to a different address; override --template-key to
pin a deterministic template key across runs.`,
	RunE: func(c *cobra.Command, _ []string) error {
		return run(c.Context())
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().String("to", "", "recipient email (defaults to openlane.auth.email)")
	command.Flags().String("template-key", "", "template key (defaults to cli-quickstart-<timestamp>)")
}

// run executes the end-to-end flow
func run(ctx context.Context) error {
	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	recipient, err := resolveRecipient()
	if err != nil {
		return err
	}

	suffix := time.Now().UTC().Format("20060102-150405")
	templateKey := cmd.Config.String("template-key")

	if templateKey == "" {
		templateKey = resourcePrefix + suffix
	}

	brandingID, err := createBranding(ctx, client, suffix)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create email branding")

		return fmt.Errorf("create branding: %w", err)
	}

	templateID, err := createTemplate(ctx, client, templateKey, suffix)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create email template")

		return fmt.Errorf("create template: %w", err)
	}

	campaignID, err := createCampaign(ctx, client, templateID, brandingID, recipient, suffix)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create campaign")

		return fmt.Errorf("create campaign: %w", err)
	}

	launch, err := client.LaunchCampaign(ctx, graphclient.LaunchCampaignInput{CampaignID: campaignID})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to launch campaign")

		return fmt.Errorf("launch campaign: %w", err)
	}

	payload := launch.LaunchCampaign
	headers := []string{"BrandingID", "TemplateID", "CampaignID", "Recipient", "Queued", "Skipped"}
	rows := [][]string{{
		brandingID,
		templateID,
		campaignID,
		recipient,
		strconv.FormatInt(payload.QueuedCount, 10),
		strconv.FormatInt(payload.SkippedCount, 10),
	}}

	return cmd.RenderTable(launch, headers, rows)
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

// createBranding creates a default email branding tagged with quickstartTag
func createBranding(ctx context.Context, client *openlaneclient.Client, suffix string) (string, error) {
	input := graphclient.CreateEmailBrandingInput{
		Name:            resourcePrefix + suffix,
		BrandName:       lo.ToPtr("Openlane Quickstart"),
		PrimaryColor:    lo.ToPtr("#1F2937"),
		BackgroundColor: lo.ToPtr("#F8FAFC"),
		TextColor:       lo.ToPtr("#0F172A"),
		LinkColor:       lo.ToPtr("#2563EB"),
		ButtonColor:     lo.ToPtr("#1F2937"),
		ButtonTextColor: lo.ToPtr("#FFFFFF"),
		IsDefault:       lo.ToPtr(true),
		Tags:            []string{quickstartTag},
	}

	resp, err := client.CreateEmailBranding(ctx, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create email branding")
		return "", err
	}

	id := resp.CreateEmailBranding.EmailBranding.ID
	logx.FromContext(ctx).Debug().Str("branding_id", id).Msg("created email branding")

	return id, nil
}

// createTemplate creates a CAMPAIGN_RECIPIENT template with the smoke-test body
func createTemplate(ctx context.Context, client *openlaneclient.Client, key, suffix string) (string, error) {
	subject := defaultSubject
	body := defaultBody

	input := graphclient.CreateEmailTemplateInput{
		Key:             key,
		Name:            "CLI Quickstart " + suffix,
		TemplateContext: enums.TemplateContextCampaignRecipient,
		SubjectTemplate: &subject,
		BodyTemplate:    &body,
		Active:          lo.ToPtr(true),
	}

	resp, err := client.CreateEmailTemplate(ctx, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", key).Msg("failed to create email template")

		return "", err
	}

	id := resp.CreateEmailTemplate.EmailTemplate.ID
	logx.FromContext(ctx).Debug().Str("template_id", id).Str("template_key", key).Msg("created email template")

	return id, nil
}

// createCampaign creates a CUSTOM campaign targeting the recipient
func createCampaign(ctx context.Context, client *openlaneclient.Client, templateID, brandingID, recipient, suffix string) (string, error) {
	campaign := &graphclient.CreateCampaignInput{
		Name:            resourcePrefix + suffix,
		Description:     lo.ToPtr("End-to-end smoke test campaign created by the integrations CLI"),
		CampaignType:    lo.ToPtr(enums.CampaignTypeCustom),
		EmailTemplateID: &templateID,
		EmailBrandingID: &brandingID,
		Tags:            []string{quickstartTag},
	}

	input := graphclient.CreateCampaignWithTargetsInput{
		Campaign: campaign,
		Targets:  []*graphclient.CreateCampaignTargetInput{{Email: recipient}},
	}

	resp, err := client.CreateCampaignWithTargets(ctx, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("recipient", recipient).Msg("failed to create campaign with targets")

		return "", err
	}

	id := resp.CreateCampaignWithTargets.Campaign.ID
	logx.FromContext(ctx).Debug().Str("campaign_id", id).Str("recipient", recipient).Msg("created campaign")

	return id, nil
}
