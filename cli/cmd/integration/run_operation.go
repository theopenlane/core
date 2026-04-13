//go:build cli

package integrations

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	api "github.com/theopenlane/core/common/openapi"
)

var runOperationCmd = &cobra.Command{
	Use:   "run-operation",
	Short: "execute a provider operation against an installed integration",
	Run: func(cmd *cobra.Command, args []string) {
		err := runOperation(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(runOperationCmd)

	runOperationCmd.Flags().StringP("integration-id", "i", "", "integration ID to run the operation against (required)")
	runOperationCmd.Flags().StringP("operation", "o", "", "operation identifier to execute (required)")
	runOperationCmd.Flags().String("op-config", "", "optional operation-specific configuration as a JSON string")
}

// validateRunOperation validates the required fields for the run-operation command
func validateRunOperation() (*api.RunIntegrationOperationRequest, error) {
	integrationID := cmd.Config.String("integration-id")
	if integrationID == "" {
		return nil, cmd.NewRequiredFieldMissingError("integration-id")
	}

	operation := cmd.Config.String("operation")
	if operation == "" {
		return nil, cmd.NewRequiredFieldMissingError("operation")
	}

	req := &api.RunIntegrationOperationRequest{
		IntegrationID: integrationID,
		Body: api.RunIntegrationOperationBody{
			Operation: operation,
		},
	}

	if cfg := cmd.Config.String("op-config"); cfg != "" {
		if !json.Valid([]byte(cfg)) {
			return nil, cmd.NewInvalidFieldError("op-config", "must be valid JSON")
		}

		req.Body.Config = json.RawMessage(cfg)
	}

	return req, nil
}

// runOperation executes or queues a provider operation against an installed integration
func runOperation(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := validateRunOperation()
	cobra.CheckErr(err)

	o, err := client.RunIntegrationOperation(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
