//go:build cli

package assessment

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing assessment",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "assessment id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the assessment")
	updateCmd.Flags().StringP("template-id", "t", "", "template ID to use for the assessment")
	updateCmd.Flags().StringP("add-platform-ids", "", "", "comma-separated list of platform IDs to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateAssessmentInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("assessment id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	templateID := cmd.Config.String("template-id")
	if templateID != "" {
		input.TemplateID = &templateID
	}

	addPlatformIDs := cmd.Config.String("add-platform-ids")
	if addPlatformIDs != "" {
		input.AddPlatformIDs = cmd.ParseIDList(addPlatformIDs)
	}

	return id, input, nil
}

// update an existing assessment in the platform
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

	o, err := client.UpdateAssessment(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
