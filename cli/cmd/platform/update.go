//go:build cli

package platform

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing platform",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "platform id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the platform")
	updateCmd.Flags().StringP("description", "d", "", "description of the platform")
	updateCmd.Flags().StringP("kind", "k", "", "kind of the platform")
	updateCmd.Flags().StringP("environment", "e", "", "environment of the platform")
	updateCmd.Flags().StringP("internal-owner", "", "", "internal owner for the platform")
	updateCmd.Flags().StringP("business-owner", "b", "", "business owner for the platform")
	updateCmd.Flags().StringP("technical-owner", "", "", "technical owner for the platform")
	updateCmd.Flags().StringP("add-asset-ids", "a", "", "comma-separated list of asset IDs to add")
	updateCmd.Flags().StringP("add-entity-ids", "", "", "comma-separated list of entity IDs to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdatePlatformInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("platform id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	kind := cmd.Config.String("kind")
	if kind != "" {
		input.PlatformKindName = &kind
	}

	environment := cmd.Config.String("environment")
	if environment != "" {
		input.EnvironmentName = &environment
	}

	internalOwner := cmd.Config.String("internal-owner")
	if internalOwner != "" {
		input.InternalOwner = &internalOwner
	}

	businessOwner := cmd.Config.String("business-owner")
	if businessOwner != "" {
		input.BusinessOwner = &businessOwner
	}

	technicalOwner := cmd.Config.String("technical-owner")
	if technicalOwner != "" {
		input.TechnicalOwner = &technicalOwner
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

// update an existing platform in the platform
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

	o, err := client.UpdatePlatform(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
