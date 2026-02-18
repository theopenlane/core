//go:build cli

package remediation

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new remediation",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("title", "t", "", "title of the remediation effort")
	createCmd.Flags().StringP("state", "s", "", "state of the remediation (e.g. PENDING, IN_PROGRESS, COMPLETED)")
	createCmd.Flags().StringP("intent", "n", "", "intent or goal of the remediation")
	createCmd.Flags().StringP("summary", "m", "", "summary of the remediation approach")
	createCmd.Flags().StringP("vulnerability-ids", "v", "", "comma-separated list of vulnerability IDs to associate with the remediation")
	createCmd.Flags().StringP("finding-ids", "f", "", "comma-separated list of finding IDs to associate with the remediation")
	createCmd.Flags().StringP("asset-ids", "", "", "comma-separated list of asset IDs to associate with the remediation")
	createCmd.Flags().StringP("entity-ids", "", "", "comma-separated list of entity IDs to associate with the remediation")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateRemediationInput, err error) {
	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	state := cmd.Config.String("state")
	if state != "" {
		input.State = &state
	}

	intent := cmd.Config.String("intent")
	if intent != "" {
		input.Intent = &intent
	}

	summary := cmd.Config.String("summary")
	if summary != "" {
		input.Summary = &summary
	}

	vulnerabilityIDs := cmd.Config.String("vulnerability-ids")
	if vulnerabilityIDs != "" {
		input.VulnerabilityIDs = cmd.ParseIDList(vulnerabilityIDs)
	}

	findingIDs := cmd.Config.String("finding-ids")
	if findingIDs != "" {
		input.FindingIDs = cmd.ParseIDList(findingIDs)
	}

	assetIDs := cmd.Config.String("asset-ids")
	if assetIDs != "" {
		input.AssetIDs = cmd.ParseIDList(assetIDs)
	}

	entityIDs := cmd.Config.String("entity-ids")
	if entityIDs != "" {
		input.EntityIDs = cmd.ParseIDList(entityIDs)
	}

	return input, nil
}

// create a new remediation
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

	o, err := client.CreateRemediation(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
