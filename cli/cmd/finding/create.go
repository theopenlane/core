//go:build cli

package finding

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new finding",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("display-name", "n", "", "display name for the finding")
	createCmd.Flags().StringP("category", "c", "", "primary category of the finding")
	createCmd.Flags().StringP("severity", "s", "", "severity label (e.g. LOW, MEDIUM, HIGH, CRITICAL)")
	createCmd.Flags().StringP("state", "t", "", "state of the finding (e.g. ACTIVE, INACTIVE)")
	createCmd.Flags().StringP("source", "o", "", "system that produced the finding")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateFindingInput, err error) {
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

	return input, nil
}

// create a new finding
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

	o, err := client.CreateFinding(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
