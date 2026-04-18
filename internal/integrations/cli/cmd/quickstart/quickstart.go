package quickstart

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
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

// defaultBody is the body template for the smoke-test email. Uses the
// CAMPAIGN_RECIPIENT variable context so the server substitutes recipient
// and company data at render time.
const defaultBody = `<html>
<body>
<p>Hi {{ .Recipient.FirstName }},</p>
<p>This is a smoke-test email from the Openlane integrations CLI for
campaign <strong>{{ .Campaign.Name }}</strong>.</p>
<p>If you received this, the end-to-end flow (branding + template + campaign
+ launch + dispatch) is working.</p>
</body>
</html>`

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
		return fmt.Errorf("create branding: %w", err)
	}

	templateID, err := createTemplate(ctx, client, templateKey, suffix)
	if err != nil {
		return fmt.Errorf("create template: %w", err)
	}

	campaignID, err := createCampaign(ctx, client, templateID, recipient, suffix)
	if err != nil {
		return fmt.Errorf("create campaign: %w", err)
	}

	launch, err := client.LaunchCampaign(ctx, graphclient.LaunchCampaignInput{CampaignID: campaignID})
	if err != nil {
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
		Name:         resourcePrefix + suffix,
		BrandName:    lo.ToPtr("Openlane Quickstart"),
		PrimaryColor: lo.ToPtr("#1F2937"),
		IsDefault:    lo.ToPtr(true),
		Tags:         []string{quickstartTag},
	}

	resp, err := client.CreateEmailBranding(ctx, input)
	if err != nil {
		return "", err
	}

	id := resp.CreateEmailBranding.EmailBranding.ID
	log.Info().Str("branding_id", id).Msg("created email branding")

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
		return "", err
	}

	id := resp.CreateEmailTemplate.EmailTemplate.ID
	log.Info().Str("template_id", id).Str("template_key", key).Msg("created email template")

	return id, nil
}

// createCampaign creates a CUSTOM campaign targeting the recipient
func createCampaign(ctx context.Context, client *openlaneclient.Client, templateID, recipient, suffix string) (string, error) {
	campaign := &graphclient.CreateCampaignInput{
		Name:         resourcePrefix + suffix,
		Description:  lo.ToPtr("End-to-end smoke test campaign created by the integrations CLI"),
		CampaignType: lo.ToPtr(enums.CampaignTypeCustom),
		TemplateID:   &templateID,
		Tags:         []string{quickstartTag},
	}

	input := graphclient.CreateCampaignWithTargetsInput{
		Campaign: campaign,
		Targets:  []*graphclient.CreateCampaignTargetInput{{Email: recipient}},
	}

	resp, err := client.CreateCampaignWithTargets(ctx, input)
	if err != nil {
		return "", err
	}

	id := resp.CreateCampaignWithTargets.Campaign.ID
	log.Info().Str("campaign_id", id).Str("recipient", recipient).Msg("created campaign")

	return id, nil
}
