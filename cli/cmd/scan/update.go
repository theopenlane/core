//go:build cli

package scan

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing scan",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "scan id to update")
	updateCmd.Flags().StringP("target", "t", "", "target of the scan")
	updateCmd.Flags().StringP("type", "y", "", "type of scan (e.g. DOMAIN, VULNERABILITY, PROVIDER)")
	updateCmd.Flags().StringP("status", "s", "", "status of the scan (e.g. PROCESSING, COMPLETED, FAILED)")
	updateCmd.Flags().StringP("environment", "e", "", "environment of the scan")
	updateCmd.Flags().StringP("assigned-to", "a", "", "who the scan is assigned to")
	updateCmd.Flags().StringP("add-vulnerability-ids", "", "", "comma-separated list of vulnerability IDs to add")
	updateCmd.Flags().StringP("add-asset-ids", "", "", "comma-separated list of asset IDs to add")
	updateCmd.Flags().StringP("add-entity-ids", "", "", "comma-separated list of entity IDs to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateScanInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("scan id")
	}

	target := cmd.Config.String("target")
	if target != "" {
		input.Target = &target
	}

	scanType := cmd.Config.String("type")
	if scanType != "" {
		input.ScanType = enums.ToScanType(scanType)
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToScanStatus(status)
	}

	environment := cmd.Config.String("environment")
	if environment != "" {
		input.EnvironmentName = &environment
	}

	assignedTo := cmd.Config.String("assigned-to")
	if assignedTo != "" {
		input.AssignedTo = &assignedTo
	}

	addVulnerabilityIDs := cmd.Config.String("add-vulnerability-ids")
	if addVulnerabilityIDs != "" {
		input.AddVulnerabilityIDs = cmd.ParseIDList(addVulnerabilityIDs)
	}

	addAssetIDs := cmd.Config.String("add-asset-ids")
	if addAssetIDs != "" {
		input.AddAssetIDs = cmd.ParseIDList(addAssetIDs)
	}

	addEntityIDs := cmd.Config.String("add-entity-ids")
	if addEntityIDs != "" {
		input.AddEntityIDs = cmd.ParseIDList(addEntityIDs)
	}

	return id, input, nil
}

// update an existing scan in the platform
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

	o, err := client.UpdateScan(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
