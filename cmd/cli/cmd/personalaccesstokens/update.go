//go:build cli

package tokens

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing personal access token",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "pat id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the personal access token")
	updateCmd.Flags().StringP("description", "d", "", "description of the pat")
	updateCmd.Flags().StringSliceP("add-organizations", "o", []string{}, "add organization(s) id to associate the pat with")
	updateCmd.Flags().StringSliceP("remove-organizations", "r", []string{}, "remove organization(s) id to associate the pat with")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdatePersonalAccessTokenInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("token id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	addOrgs := cmd.Config.Strings("add-organizations")
	if addOrgs != nil {
		input.AddOrganizationIDs = addOrgs
	}

	removeOrgs := cmd.Config.Strings("remove-organizations")
	if removeOrgs != nil {
		input.RemoveOrganizationIDs = removeOrgs
	}

	return id, input, nil
}

// update an existing personal access token
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

	o, err := client.UpdatePersonalAccessToken(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
