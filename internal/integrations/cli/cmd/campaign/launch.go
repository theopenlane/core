//go:build examples

package campaign

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "launch a campaign and dispatch emails to its targets",
	Long: `Launch dispatches the campaign through the configured email integration.
Pass --resend to include targets that already received a send, and
--scheduled-at (YYYY-MM-DD or ISO8601) to defer the dispatch.`,
	RunE: func(c *cobra.Command, _ []string) error {
		return launch(c.Context())
	},
}

func init() {
	command.AddCommand(launchCmd)

	launchCmd.Flags().StringP("campaign-id", "c", "", "campaign ID to launch (required)")
	launchCmd.Flags().Bool("resend", false, "resend to previously-sent targets")
	launchCmd.Flags().String("scheduled-at", "", "schedule the launch for a future time (YYYY-MM-DD or ISO8601)")
}

// buildLaunchInput builds the LaunchCampaignInput from config + flags
func buildLaunchInput() (graphclient.LaunchCampaignInput, error) {
	var input graphclient.LaunchCampaignInput

	campaignID := cmd.Config.String("campaign-id")
	if campaignID == "" {
		return input, ErrCampaignIDRequired
	}

	input.CampaignID = campaignID

	resend := cmd.Config.Bool("resend")
	input.Resend = &resend

	if scheduled := cmd.Config.String("scheduled-at"); scheduled != "" {
		dt, err := models.ToDateTime(scheduled)
		if err != nil {
			return input, err
		}

		input.ScheduledAt = dt
	}

	return input, nil
}

// launch executes the LaunchCampaign mutation and prints a dispatch summary
func launch(ctx context.Context) error {
	input, err := buildLaunchInput()
	if err != nil {
		return err
	}

	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	resp, err := client.LaunchCampaign(ctx, input)
	if err != nil {
		return err
	}

	payload := resp.LaunchCampaign

	headers := []string{"CampaignID", "Name", "Status", "Queued", "Skipped"}
	rows := [][]string{{
		payload.Campaign.ID,
		payload.Campaign.Name,
		string(payload.Campaign.Status),
		strconv.FormatInt(payload.QueuedCount, 10),
		strconv.FormatInt(payload.SkippedCount, 10),
	}}

	return cmd.RenderTable(resp, headers, rows)
}
