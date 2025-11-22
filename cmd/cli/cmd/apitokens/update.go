//go:build cli

package apitokens

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an api token token",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "api token id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the api token token")
	updateCmd.Flags().StringP("description", "d", "", "description of the api token")
	updateCmd.Flags().StringSlice("scopes", []string{}, "scopes to add to the api token"+scopeFlagConfig())
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateAPITokenInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("token id")
	}

	// Craft update input
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	scopes := cmd.Config.Strings("scopes")
	if len(scopes) > 0 {
		input.Scopes = scopes
	}

	return id, input, nil
}

// update an existing api token
func update(ctx context.Context) error {
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateAPIToken(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
