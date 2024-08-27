package group

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	cmd.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the group")
	createCmd.Flags().StringP("display-name", "s", "", "display name of the group")
	createCmd.Flags().StringP("description", "d", "", "description of the group")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateGroupInput, err error) {
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

	return input, nil
}

// create a new group
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateGroup(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
