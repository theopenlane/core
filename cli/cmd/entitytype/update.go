//go:build cli

package entitytype

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing entity type",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "entity type id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the entity type")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateEntityTypeInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("entity type id")
	}

	// validation of required fields for the update command
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	return id, input, nil
}

// update an existing entity type in the platform
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

	o, err := client.UpdateEntityType(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
