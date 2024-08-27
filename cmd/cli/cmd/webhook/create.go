package webhooks

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new webhook",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the webhook")
	createCmd.Flags().StringP("description", "d", "", "description of the webhook")
	createCmd.Flags().StringP("url", "u", "", "the destination url the webhook is sent to")
	createCmd.Flags().BoolP("enabled", "e", true, "if the webhook is enabled")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateWebhookInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.DestinationURL = cmd.Config.String("url")
	if input.DestinationURL == "" {
		return input, cmd.NewRequiredFieldMissingError("url")
	}

	enabled := cmd.Config.Bool("enabled")
	if !enabled {
		input.Enabled = &enabled
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	return input, nil
}

// create a new webhook
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateWebhook(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
