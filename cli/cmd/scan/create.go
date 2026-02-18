//go:build cli

package scan

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new scan",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("target", "t", "", "target of the scan (e.g. domain name, IP address, codebase)")
	createCmd.Flags().StringP("type", "y", "", "type of scan (e.g. DOMAIN, VULNERABILITY, PROVIDER)")
	createCmd.Flags().StringP("environment", "e", "", "environment of the scan")
	createCmd.Flags().StringP("assigned-to", "a", "", "who the scan is assigned to")
	createCmd.Flags().StringP("vulnerability-ids", "v", "", "comma-separated list of vulnerability IDs to associate with the scan")
	createCmd.Flags().StringP("asset-ids", "", "", "comma-separated list of asset IDs to associate with the scan")
	createCmd.Flags().StringP("entity-ids", "", "", "comma-separated list of entity IDs to associate with the scan")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateScanInput, err error) {
	input.Target = cmd.Config.String("target")
	if input.Target == "" {
		return input, cmd.NewRequiredFieldMissingError("target")
	}

	scanType := cmd.Config.String("type")
	if scanType != "" {
		input.ScanType = enums.ToScanType(scanType)
	}

	environment := cmd.Config.String("environment")
	if environment != "" {
		input.EnvironmentName = &environment
	}

	assignedTo := cmd.Config.String("assigned-to")
	if assignedTo != "" {
		input.AssignedTo = &assignedTo
	}

	vulnerabilityIDs := cmd.Config.String("vulnerability-ids")
	if vulnerabilityIDs != "" {
		input.VulnerabilityIDs = cmd.ParseIDList(vulnerabilityIDs)
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

// create a new scan
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

	o, err := client.CreateScan(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
