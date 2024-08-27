package integrations

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var integrationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new integration",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(integrationCreateCmd)

	integrationCreateCmd.Flags().StringP("name", "n", "", "name of the integration")
	integrationCreateCmd.Flags().StringP("description", "d", "", "description of the integration")
	integrationCreateCmd.Flags().StringP("kind", "k", "", "the kind of integration")
	integrationCreateCmd.Flags().StringP("webhook-id", "w", "", "the webhook id to associate with the integration")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateIntegrationInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	kind := cmd.Config.String("kind")
	if kind == "" {
		return input, cmd.NewRequiredFieldMissingError("kind")
	}

	input.Kind = &kind

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	webhookID := cmd.Config.String("webhook-id")
	if webhookID != "" {
		input.WebhookIDs = append(input.WebhookIDs, webhookID)
	}

	return input, nil
}

// create an integration in the platform
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateIntegration(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
