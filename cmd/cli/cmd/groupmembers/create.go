//go:build cli

package groupmembers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "add a user to a group",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("group-id", "g", "", "group id")
	createCmd.Flags().StringP("user-id", "u", "", "user id")
	createCmd.Flags().StringP("role", "r", "member", "role to assign the user (member, admin)")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateGroupMembershipInput, err error) {
	input.GroupID = cmd.Config.String("group-id")
	if input.GroupID == "" {
		return input, cmd.NewRequiredFieldMissingError("group id")
	}

	input.UserID = cmd.Config.String("user-id")
	if input.UserID == "" {
		return input, cmd.NewRequiredFieldMissingError("user id")
	}

	// role defaults to `member` so it is not required
	role := cmd.Config.String("role")
	if role != "" {
		r, err := cmd.GetRoleEnum(role)
		cobra.CheckErr(err)

		input.Role = &r
	}

	return input, nil
}

// create adds a user to a group in the platform
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

	o, err := client.CreateGroupMembership(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
