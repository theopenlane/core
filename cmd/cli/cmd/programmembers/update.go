package programmembers

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a user's role in a program",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("program-id", "p", "", "program id")
	updateCmd.Flags().StringP("user-id", "u", "", "user id")
	updateCmd.Flags().StringP("role", "r", "member", "role to assign the user (member, admin)")
}

// updateValidation validates the required fields for the command
func updateValidation() (where openlaneclient.ProgramMembershipWhereInput, input openlaneclient.UpdateProgramMembershipInput, err error) {
	programID := cmd.Config.String("program-id")
	if programID == "" {
		return where, input, cmd.NewRequiredFieldMissingError("program id")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return where, input, cmd.NewRequiredFieldMissingError("user id")
	}

	role := cmd.Config.String("role")
	if role == "" {
		return where, input, cmd.NewRequiredFieldMissingError("role")
	}

	r, err := cmd.GetRoleEnum(role)
	cobra.CheckErr(err)

	where = openlaneclient.ProgramMembershipWhereInput{
		ProgramID: &programID,
		UserID:    &userID,
	}

	input = openlaneclient.UpdateProgramMembershipInput{
		Role: &r,
	}

	return where, input, nil
}

// update a user's role in a program in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	where, input, err := updateValidation()
	cobra.CheckErr(err)

	programMembers, err := client.GetProgramMembersByProgramID(ctx, &where)
	cobra.CheckErr(err)

	if len(programMembers.ProgramMemberships.Edges) != 1 {
		return errors.New("error getting existing relation") //nolint:err113
	}

	o, err := client.UpdateUserRoleInProgram(ctx, programMembers.ProgramMemberships.Edges[0].Node.ID, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
