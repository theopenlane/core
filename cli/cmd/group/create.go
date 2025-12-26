//go:build cli

package group

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new group",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the group")
	createCmd.Flags().StringP("display-name", "s", "", "display name of the group")
	createCmd.Flags().StringP("description", "d", "", "description of the group")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the group")
	createCmd.Flags().BoolP("private", "p", false, "set group to private")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateGroupInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("group name")
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	if cmd.Config.Bool("private") {
		input.CreateGroupSettings = &graphclient.CreateGroupSettingInput{
			Visibility: &enums.VisibilityPrivate,
		}
	}

	return input, nil
}

// create a new group
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

	o, err := client.CreateGroup(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
