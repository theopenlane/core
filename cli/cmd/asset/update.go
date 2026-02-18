//go:build cli

package asset

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing asset",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "asset id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the asset")
	updateCmd.Flags().StringP("description", "d", "", "description of the asset")
	updateCmd.Flags().StringP("type", "t", "", "type of asset (e.g. TECHNOLOGY, DOMAIN, DEVICE)")
	updateCmd.Flags().StringP("identifier", "", "", "unique identifier like domain, device id, etc")
	updateCmd.Flags().StringP("website", "w", "", "website of the asset, if applicable")
	updateCmd.Flags().StringP("physical-location", "", "", "physical location of the asset")
	updateCmd.Flags().StringP("region", "r", "", "region where the asset operates or is hosted")
	updateCmd.Flags().BoolP("contains-pii", "", false, "whether the asset stores or processes PII")
	updateCmd.Flags().StringP("environment", "e", "", "environment of the asset")
	updateCmd.Flags().StringP("add-platform-ids", "p", "", "comma-separated list of platform IDs to add")
	updateCmd.Flags().StringP("add-entity-ids", "", "", "comma-separated list of entity IDs to add")
	updateCmd.Flags().StringP("add-scan-ids", "", "", "comma-separated list of scan IDs to add")
	updateCmd.Flags().StringP("add-control-ids", "", "", "comma-separated list of control IDs to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateAssetInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("asset id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	assetType := cmd.Config.String("type")
	if assetType != "" {
		input.AssetType = enums.ToAssetType(assetType)
	}

	identifier := cmd.Config.String("identifier")
	if identifier != "" {
		input.Identifier = &identifier
	}

	website := cmd.Config.String("website")
	if website != "" {
		input.Website = &website
	}

	physicalLocation := cmd.Config.String("physical-location")
	if physicalLocation != "" {
		input.PhysicalLocation = &physicalLocation
	}

	region := cmd.Config.String("region")
	if region != "" {
		input.Region = &region
	}

	containsPii := cmd.Config.Bool("contains-pii")
	input.ContainsPii = &containsPii

	environment := cmd.Config.String("environment")
	if environment != "" {
		input.EnvironmentName = &environment
	}

	addPlatformIDs := cmd.Config.String("add-platform-ids")
	if addPlatformIDs != "" {
		input.AddPlatformIDs = cmd.ParseIDList(addPlatformIDs)
	}

	addEntityIDs := cmd.Config.String("add-entity-ids")
	if addEntityIDs != "" {
		input.AddEntityIDs = cmd.ParseIDList(addEntityIDs)
	}

	addScanIDs := cmd.Config.String("add-scan-ids")
	if addScanIDs != "" {
		input.AddScanIDs = cmd.ParseIDList(addScanIDs)
	}

	addControlIDs := cmd.Config.String("add-control-ids")
	if addControlIDs != "" {
		input.AddControlIDs = cmd.ParseIDList(addControlIDs)
	}

	return id, input, nil
}

// update an existing asset in the platform
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

	o, err := client.UpdateAsset(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
