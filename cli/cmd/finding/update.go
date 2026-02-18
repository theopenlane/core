//go:build cli

package finding

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing finding",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "finding id to update")
	updateCmd.Flags().StringP("display-name", "n", "", "display name for the finding")
	updateCmd.Flags().StringP("category", "c", "", "primary category of the finding")
	updateCmd.Flags().StringP("severity", "s", "", "severity label (e.g. LOW, MEDIUM, HIGH, CRITICAL)")
	updateCmd.Flags().StringP("state", "t", "", "state of the finding (e.g. ACTIVE, INACTIVE)")
	updateCmd.Flags().StringP("source", "o", "", "system that produced the finding")
	updateCmd.Flags().StringP("add-vulnerability-ids", "", "", "comma-separated list of vulnerability IDs to add")
	updateCmd.Flags().StringP("add-remediation-ids", "", "", "comma-separated list of remediation IDs to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateFindingInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("finding id")
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	category := cmd.Config.String("category")
	if category != "" {
		input.Category = &category
	}

	severity := cmd.Config.String("severity")
	if severity != "" {
		input.Severity = &severity
	}

	state := cmd.Config.String("state")
	if state != "" {
		input.State = &state
	}

	source := cmd.Config.String("source")
	if source != "" {
		input.Source = &source
	}

	addVulnerabilityIDs := cmd.Config.String("add-vulnerability-ids")
	if addVulnerabilityIDs != "" {
		input.AddVulnerabilityIDs = cmd.ParseIDList(addVulnerabilityIDs)
	}

	addRemediationIDs := cmd.Config.String("add-remediation-ids")
	if addRemediationIDs != "" {
		input.AddRemediationIDs = cmd.ParseIDList(addRemediationIDs)
	}

	return id, input, nil
}

// update an existing finding in the platform
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

	o, err := client.UpdateFinding(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
