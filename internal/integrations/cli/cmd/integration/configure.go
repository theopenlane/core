package integration

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	api "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "configure a new integration or update an existing integration's credentials",
	Long: `Configure creates (or updates) an integration installation for the
organization resolved from the authenticated session.

--body carries the provider-specific credential JSON (required on create);
--user-input carries the installation-scoped configuration JSON. Either flag
may be a raw JSON string or an @path/to/file.json reference.

Example (email / resend):
  integrations integration configure \
    --definition-id def_01EMAILINT00000000000000001 \
    --credential-ref email-api-key \
    --body '{"apiKey":"re_xxx","provider":"resend"}' \
    --user-input '{"fromEmail":"noreply@example.com","companyName":"Acme"}'`,
	RunE: func(c *cobra.Command, _ []string) error {
		return configure(c.Context())
	},
}

func init() {
	command.AddCommand(configureCmd)

	configureCmd.Flags().StringP("definition-id", "d", "", "integration definition ID (required)")
	configureCmd.Flags().String("integration-id", "", "existing integration ID to update; omit to create a new installation")
	configureCmd.Flags().String("credential-ref", "", "credential slot identifier for this configuration")
	configureCmd.Flags().String("body", "", "provider-specific credential fields as JSON (or @file.json)")
	configureCmd.Flags().String("user-input", "", "optional installation-scoped configuration as JSON (or @file.json)")
}

// configure dispatches the configure request
func configure(ctx context.Context) error {
	req, err := buildConfigureRequest()
	if err != nil {
		return err
	}

	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	resp, err := client.ConfigureIntegration(ctx, req)
	if err != nil {
		return err
	}

	rows := [][]string{{
		resp.Provider,
		resp.IntegrationID,
		resp.HealthStatus,
		resp.HealthSummary,
		resp.WebhookEndpointURL,
	}}

	headers := []string{"Provider", "IntegrationID", "HealthStatus", "HealthSummary", "WebhookEndpointURL"}

	return cmd.RenderTable(resp, headers, rows)
}

// buildConfigureRequest validates flags and builds the REST request payload
func buildConfigureRequest() (*api.ConfigureIntegrationRequest, error) {
	definitionID := cmd.Config.String("definition-id")
	if definitionID == "" {
		return nil, ErrDefinitionIDRequired
	}

	req := &api.ConfigureIntegrationRequest{
		DefinitionID:  definitionID,
		IntegrationID: cmd.Config.String("integration-id"),
		CredentialRef: cmd.Config.String("credential-ref"),
	}

	body, err := loadJSONFlag(cmd.Config.String("body"))
	if err != nil {
		return nil, ErrInvalidBody
	}

	req.Body = body

	userInput, err := loadJSONFlag(cmd.Config.String("user-input"))
	if err != nil {
		return nil, ErrInvalidUserInput
	}

	req.UserInput = userInput

	return req, nil
}

// loadJSONFlag returns the raw JSON for a flag value. Values prefixed with `@`
// are read from disk; anything else is validated and returned verbatim. Empty
// values pass through as nil.
func loadJSONFlag(value string) (json.RawMessage, error) {
	if value == "" {
		return nil, nil
	}

	if value[0] == '@' {
		data, err := os.ReadFile(value[1:])
		if err != nil {
			return nil, err
		}

		if !json.Valid(data) {
			return nil, ErrInvalidJSON
		}

		return data, nil
	}

	if !json.Valid([]byte(value)) {
		return nil, ErrInvalidJSON
	}

	return json.RawMessage(value), nil
}
