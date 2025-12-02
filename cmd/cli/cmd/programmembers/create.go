//go:build cli

package programmembers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "add a user to a program",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("program-id", "p", "", "program id")
	createCmd.Flags().StringP("user-id", "u", "", "user id")
	createCmd.Flags().StringP("role", "r", "member", "role to assign the user (member, admin)")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateProgramMembershipInput, err error) {
	input.ProgramID = cmd.Config.String("program-id")
	if input.ProgramID == "" {
		return input, cmd.NewRequiredFieldMissingError("program id")
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

// create adds a user to a program in the platform
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

	o, err := client.CreateProgramMembership(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
