//go:build cli

package directorygroup

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing directory group",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "directory group id to update")
	updateCmd.Flags().String("display-name", "", "display name of the group")
	updateCmd.Flags().String("email", "", "primary group email address")
	updateCmd.Flags().StringP("description", "d", "", "free-form description")
	updateCmd.Flags().StringSlice("tags", []string{}, "tags associated with the group")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateDirectoryGroupInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("directory group id")
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	email := cmd.Config.String("email")
	if email != "" {
		input.Email = &email
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return id, input, nil
}

// update an existing directory group
func update(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateDirectoryGroup(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
