package groupmembers

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove a user from a group",
	Run: func(cmd *cobra.Command, args []string) {
		err := delete(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("group-id", "g", "", "group id")
	deleteCmd.Flags().StringP("user-id", "u", "", "user id")
}

// deleteValidation validates the required fields for the command
func deleteValidation() (where openlaneclient.GroupMembershipWhereInput, err error) {
	groupID := cmd.Config.String("group-id")
	if groupID == "" {
		return where, cmd.NewRequiredFieldMissingError("group id")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return where, cmd.NewRequiredFieldMissingError("user id")
	}

	where.GroupID = &groupID
	where.UserID = &userID

	return where, nil
}

// delete removes a user from a group in the platform
func delete(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	where, err := deleteValidation()
	cobra.CheckErr(err)

	groupMembers, err := client.GetGroupMembersByGroupID(ctx, &where)
	cobra.CheckErr(err)

	if len(groupMembers.GroupMemberships.Edges) != 1 {
		return errors.New("error getting existing relation") //nolint:err113
	}

	o, err := client.RemoveUserFromGroup(ctx, groupMembers.GroupMemberships.Edges[0].Node.ID)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
