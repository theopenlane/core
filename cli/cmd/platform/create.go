//go:build cli

package platform

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new platform",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the platform")
	createCmd.Flags().StringP("description", "d", "", "description of the platform")
	createCmd.Flags().StringP("kind", "k", "", "kind of the platform")
	createCmd.Flags().StringP("environment", "e", "", "environment of the platform")
	createCmd.Flags().StringP("internal-owner", "i", "", "internal owner for the platform")
	createCmd.Flags().StringP("business-owner", "b", "", "business owner for the platform")
	createCmd.Flags().StringP("technical-owner", "", "", "technical owner for the platform")
	createCmd.Flags().StringP("asset-ids", "a", "", "comma-separated list of asset IDs to associate with the platform")
	createCmd.Flags().StringP("entity-ids", "", "", "comma-separated list of entity IDs to associate with the platform")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreatePlatformInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
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

// create a new platform
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

	o, err := client.CreatePlatform(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
