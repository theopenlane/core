package integrations

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new integration",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the integration")
	createCmd.Flags().StringP("description", "d", "", "description of the integration")
	createCmd.Flags().StringP("kind", "k", "", "the kind of integration")
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

	return input, nil
}

// create an integration in the platform
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

	o, err := client.CreateIntegration(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
