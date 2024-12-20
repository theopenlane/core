package groupmembers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get existing members of a group",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("group-id", "g", "", "group id to query")
}

// get existing members of a group
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	// filter options
	id := cmd.Config.String("group-id")
	if id == "" {
		return cmd.NewRequiredFieldMissingError("group id")
	}

	where := openlaneclient.GroupMembershipWhereInput{
		GroupID: &id,
	}

	o, err := client.GetGroupMembersByGroupID(ctx, &where)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
