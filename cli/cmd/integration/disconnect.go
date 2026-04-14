//go:build cli

package integrations

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	api "github.com/theopenlane/core/common/openapi"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "disconnect an installed integration",
	Run: func(cmd *cobra.Command, args []string) {
		err := disconnect(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(disconnectCmd)

	disconnectCmd.Flags().StringP("integration-id", "i", "", "integration ID to disconnect (required)")
}

// validateDisconnect validates the required fields for the disconnect command
func validateDisconnect() (*api.DisconnectIntegrationRequest, error) {
	integrationID := cmd.Config.String("integration-id")
	if integrationID == "" {
		return nil, cmd.NewRequiredFieldMissingError("integration-id")
	}

	return &api.DisconnectIntegrationRequest{
		IntegrationID: integrationID,
	}, nil
}

// disconnect tears down an installed integration
func disconnect(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := validateDisconnect()
	cobra.CheckErr(err)

	o, err := client.DisconnectIntegration(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
