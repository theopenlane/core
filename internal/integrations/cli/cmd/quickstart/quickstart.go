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
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/logx"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/go-client/graphclient"
)

// quickstartTag marks resources created by this command so they are easy to
// spot (and clean up) later
const quickstartTag = "cli-quickstart"

// resourcePrefix is prepended to generated resource names
const resourcePrefix = "cli-quickstart-"

// defaultsJSON is the canonical BrandedMessageRequest payload used to seed the
// quickstart template. Edit defaults.example.json to change the wording
//
//go:embed defaults.example.json
var defaultsJSON []byte

var command = &cobra.Command{
	Use:   "quickstart",
	Short: "end-to-end Openlane Campaigns smoke test: create branding + template + campaign, then launch",
	Long: `Provisions an Openlane Campaigns pipeline end-to-end: a branded, branded-message
template seeded from defaults.example.json, a campaign targeting the authenticated
user, and a launch that exercises dispatch through the email provider.

All resources are tagged "cli-quickstart" for easy discovery and cleanup. Branding
is attached to the template so the campaign inherits it transparently.

Override --to to send to a different address.`,
	RunE: func(c *cobra.Command, _ []string) error {
		return run(c.Context())
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().String("to", "", "recipient email (defaults to openlane.auth.email)")
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
	templateKey := email.BrandedMessageOp.Name()

	templateID, err := createTemplate(ctx, client, templateKey, suffix)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create email template")

		return fmt.Errorf("create template: %w", err)
	}

	campaignID, err := createCampaign(ctx, client, templateID, recipient, suffix)
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

// loadDefaults unmarshals the embedded defaults.example.json into the map shape
// expected by graphclient.CreateEmailTemplateInput.Defaults
func loadDefaults() (map[string]any, error) {
	out := map[string]any{}
	if err := json.Unmarshal(defaultsJSON, &out); err != nil {
		return nil, fmt.Errorf("parse quickstart defaults: %w", err)
	}

	return out, nil
}

// createTemplate creates a CAMPAIGN_RECIPIENT template bound to the branded-message catalog entry.
// The Defaults payload pre-fills the BrandedMessageRequest fields; recipient/campaign context is
// layered on by the dispatcher at send time
func createTemplate(ctx context.Context, client *openlaneclient.Client, key, suffix string) (string, error) {
	defaults, err := loadDefaults()
	if err != nil {
		return "", err
	}

	input := graphclient.CreateEmailTemplateInput{
		Key:             key,
		Name:            "Openlane Campaigns quickstart " + suffix,
		TemplateContext: enums.TemplateContextCampaignRecipient,
		Defaults:        defaults,
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

// createCampaign creates a CUSTOM campaign targeting the recipient. Branding is attached via the
// template's EmailBrandingIDs edge, so the campaign does not set EmailBrandingID directly
func createCampaign(ctx context.Context, client *openlaneclient.Client, templateID, recipient, suffix string) (string, error) {
	campaign := &graphclient.CreateCampaignInput{
		Name:            resourcePrefix + suffix,
		Description:     lo.ToPtr("Openlane Campaigns end-to-end smoke test"),
		CampaignType:    lo.ToPtr(enums.CampaignTypeCustom),
		EmailTemplateID: &templateID,
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
