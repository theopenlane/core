package programmembers

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove a user from a program",
	Run: func(cmd *cobra.Command, args []string) {
		err := delete(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("program-id", "g", "", "program id")
	deleteCmd.Flags().StringP("user-id", "u", "", "user id")
}

// deleteValidation validates the required fields for the command
func deleteValidation() (where openlaneclient.ProgramMembershipWhereInput, err error) {
	programID := cmd.Config.String("program-id")
	if programID == "" {
		return where, cmd.NewRequiredFieldMissingError("program id")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return where, cmd.NewRequiredFieldMissingError("user id")
	}

	where.ProgramID = &programID
	where.UserID = &userID

	return where, nil
}

// delete removes a user from a program in the platform
func delete(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	where, err := deleteValidation()
	cobra.CheckErr(err)

	programMembers, err := client.GetProgramMembersByProgramID(ctx, &where)
	cobra.CheckErr(err)

	if len(programMembers.ProgramMemberships.Edges) != 1 {
		return errors.New("error getting existing relation") //nolint:err113
	}

	o, err := client.RemoveUserFromProgram(ctx, programMembers.ProgramMemberships.Edges[0].Node.ID)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
