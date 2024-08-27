package orgmembers

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove a user from a organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := delete(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("user-id", "u", "", "user id")
}

// deleteValidation validates the required fields for the command
func deleteValidation() (where openlaneclient.OrgMembershipWhereInput, err error) {
	uID := cmd.Config.String("user-id")
	if uID == "" {
		return where, cmd.NewRequiredFieldMissingError("user id")
	}

	where.UserID = &uID

	return where, nil
}

func delete(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	where, err := deleteValidation()
	cobra.CheckErr(err)

	orgMembers, err := client.GetOrgMembersByOrgID(ctx, &where)
	cobra.CheckErr(err)

	if len(orgMembers.OrgMemberships.Edges) != 1 {
		cobra.CheckErr(errors.New("error getting existing relation")) //nolint:err113
	}

	id := orgMembers.OrgMemberships.Edges[0].Node.ID

	o, err := client.RemoveUserFromOrg(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
