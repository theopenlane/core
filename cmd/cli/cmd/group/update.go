//go:build cli

package group

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing group",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "group id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the group")
	updateCmd.Flags().StringP("display-name", "s", "", "display name of the group")
	updateCmd.Flags().StringP("description", "d", "", "description of the group")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateGroupInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("group id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	return id, input, nil
}

// update an existing group in the platform
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

	o, err := client.UpdateGroup(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
