//go:build cli

package groupmembers

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a user's role in a group",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("group-id", "g", "", "group id")
	updateCmd.Flags().StringP("user-id", "u", "", "user id")
	updateCmd.Flags().StringP("role", "r", "member", "role to assign the user (member, admin)")
}

// updateValidation validates the required fields for the command
func updateValidation() (where openlaneclient.GroupMembershipWhereInput, input openlaneclient.UpdateGroupMembershipInput, err error) {
	groupID := cmd.Config.String("group-id")
	if groupID == "" {
		return where, input, cmd.NewRequiredFieldMissingError("group id")
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

	where = openlaneclient.GroupMembershipWhereInput{
		GroupID: &groupID,
		UserID:  &userID,
	}

	input = openlaneclient.UpdateGroupMembershipInput{
		Role: &r,
	}

	return where, input, nil
}

// update a user's role in a group in the platform
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

	groupMembers, err := client.GetGroupMemberships(ctx, cmd.First, cmd.Last, &where)
	cobra.CheckErr(err)

	if len(groupMembers.GroupMemberships.Edges) != 1 {
		return errors.New("error getting existing relation") //nolint:err113
	}

	o, err := client.UpdateGroupMembership(ctx, groupMembers.GroupMemberships.Edges[0].Node.ID, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
