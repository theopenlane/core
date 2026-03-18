//go:build cli

package campaign

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new campaign",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the campaign")
	createCmd.Flags().StringP("description", "d", "", "description of the campaign")
	createCmd.Flags().StringP("type", "y", "", "type of campaign (QUESTIONNAIRE, TRAINING, POLICY_ATTESTATION, VENDOR_ASSESSMENT, CUSTOM)")
	createCmd.Flags().StringP("status", "s", "", "status of the campaign")
	createCmd.Flags().StringP("assessment-id", "a", "", "assessment ID to associate with the campaign")
	createCmd.Flags().StringP("template-id", "t", "", "template ID to use for the campaign")
	createCmd.Flags().BoolP("is-recurring", "r", false, "whether the campaign recurs on a schedule")
	createCmd.Flags().StringP("recurrence-frequency", "f", "", "recurrence cadence (YEARLY, QUARTERLY, BIANNUALLY, MONTHLY)")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateCampaignInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
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

	assessmentID := cmd.Config.String("assessment-id")
	if assessmentID != "" {
		input.AssessmentID = &assessmentID
	}

	templateID := cmd.Config.String("template-id")
	if templateID != "" {
		input.TemplateID = &templateID
	}

	isRecurring := cmd.Config.Bool("is-recurring")
	input.IsRecurring = &isRecurring

	recurrenceFrequency := cmd.Config.String("recurrence-frequency")
	if recurrenceFrequency != "" {
		input.RecurrenceFrequency = enums.ToFrequency(recurrenceFrequency)
	}

	return input, nil
}

// create a new campaign
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateCampaign(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
