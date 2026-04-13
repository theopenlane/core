//go:build cli

package integrations

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	api "github.com/theopenlane/core/common/openapi"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "configure a new integration or update an existing integration's config",
	Run: func(cmd *cobra.Command, args []string) {
		err := configure(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(configureCmd)

	configureCmd.Flags().StringP("definition-id", "d", "", "integration definition ID (required)")
	configureCmd.Flags().String("integration-id", "", "existing integration ID to update; omit to create new")
	configureCmd.Flags().String("credential-ref", "", "credential slot to configure")
	configureCmd.Flags().String("body", "", "provider-specific credential fields as a JSON string")
	configureCmd.Flags().String("user-input", "", "optional installation-scoped provider configuration as a JSON string")
}

// validateConfigure validates the required fields for the configure command
func validateConfigure() (*api.ConfigureIntegrationRequest, error) {
	definitionID := cmd.Config.String("definition-id")
	if definitionID == "" {
		return nil, cmd.NewRequiredFieldMissingError("definition-id")
	}

	req := &api.ConfigureIntegrationRequest{
		DefinitionID:  definitionID,
		IntegrationID: cmd.Config.String("integration-id"),
		CredentialRef: cmd.Config.String("credential-ref"),
	}

	if body := cmd.Config.String("body"); body != "" {
		if !json.Valid([]byte(body)) {
			return nil, cmd.NewInvalidFieldError("body", "must be valid JSON")
		}

		req.Body = json.RawMessage(body)
	}

	if userInput := cmd.Config.String("user-input"); userInput != "" {
		if !json.Valid([]byte(userInput)) {
			return nil, cmd.NewInvalidFieldError("user-input", "must be valid JSON")
		}

		req.UserInput = json.RawMessage(userInput)
	}

	return req, nil
}

// configure creates or updates an integration configuration
func configure(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := validateConfigure()
	cobra.CheckErr(err)

	o, err := client.ConfigureIntegration(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
