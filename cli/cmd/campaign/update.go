//go:build cli

package campaign

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing campaign",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "campaign id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the campaign")
	updateCmd.Flags().StringP("description", "d", "", "description of the campaign")
	updateCmd.Flags().StringP("type", "y", "", "type of campaign (QUESTIONNAIRE, TRAINING, POLICY_ATTESTATION, VENDOR_ASSESSMENT, CUSTOM)")
	updateCmd.Flags().StringP("status", "s", "", "status of the campaign")
	updateCmd.Flags().BoolP("is-recurring", "r", false, "whether the campaign recurs on a schedule")
	updateCmd.Flags().StringP("recurrence-frequency", "f", "", "recurrence cadence (YEARLY, QUARTERLY, BIANNUALLY, MONTHLY)")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateCampaignInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("campaign id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	campaignType := cmd.Config.String("type")
	if campaignType != "" {
		input.CampaignType = enums.ToCampaignType(campaignType)
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToCampaignStatus(status)
	}

	recurrenceFrequency := cmd.Config.String("recurrence-frequency")
	if recurrenceFrequency != "" {
		input.RecurrenceFrequency = enums.ToFrequency(recurrenceFrequency)

		isRecurring := cmd.Config.Bool("is-recurring")
		input.IsRecurring = &isRecurring
	}

	return id, input, nil
}

// update an existing campaign in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateCampaign(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
