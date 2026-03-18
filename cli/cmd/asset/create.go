//go:build cli

package asset

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new asset",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the asset")
	createCmd.Flags().StringP("description", "d", "", "description of the asset")
	createCmd.Flags().StringP("type", "t", "", "type of asset (e.g. TECHNOLOGY, DOMAIN, DEVICE)")
	createCmd.Flags().StringP("identifier", "", "", "unique identifier like domain, device id, etc")
	createCmd.Flags().StringP("website", "w", "", "website of the asset, if applicable")
	createCmd.Flags().StringP("physical-location", "", "", "physical location of the asset")
	createCmd.Flags().StringP("region", "r", "", "region where the asset operates or is hosted")
	createCmd.Flags().BoolP("contains-pii", "", false, "whether the asset stores or processes PII")
	createCmd.Flags().StringP("environment", "e", "", "environment of the asset")
	createCmd.Flags().StringP("platform-ids", "p", "", "comma-separated list of platform IDs to associate with the asset")
	createCmd.Flags().StringP("entity-ids", "", "", "comma-separated list of entity IDs to associate with the asset")
	createCmd.Flags().StringP("scan-ids", "", "", "comma-separated list of scan IDs to associate with the asset")
	createCmd.Flags().StringP("control-ids", "", "", "comma-separated list of control IDs to associate with the asset")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateAssetInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
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

	platformIDs := cmd.Config.String("platform-ids")
	if platformIDs != "" {
		input.PlatformIDs = cmd.ParseIDList(platformIDs)
	}

	entityIDs := cmd.Config.String("entity-ids")
	if entityIDs != "" {
		input.EntityIDs = cmd.ParseIDList(entityIDs)
	}

	scanIDs := cmd.Config.String("scan-ids")
	if scanIDs != "" {
		input.ScanIDs = cmd.ParseIDList(scanIDs)
	}

	controlIDs := cmd.Config.String("control-ids")
	if controlIDs != "" {
		input.ControlIDs = cmd.ParseIDList(controlIDs)
	}

	return input, nil
}

// create a new asset
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

	o, err := client.CreateAsset(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
