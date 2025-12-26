//go:build cli

package mappabledomain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new mappable domain",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the mappable domain")
	createCmd.Flags().StringP("zone-id", "", "", "ID of the dns zone")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "tags for the mappable domain")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateMappableDomainInput, err error) {
	name := cmd.Config.String("name")
	if name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	zoneID := cmd.Config.String("zone-id")
	if zoneID == "" {
		return input, cmd.NewRequiredFieldMissingError("zone-id")
	}

	input.Name = name
	input.ZoneID = zoneID

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new mappable domain
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

	o, err := client.CreateMappableDomain(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
