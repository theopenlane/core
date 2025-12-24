//go:build cli

package groupmembers

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
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

	order := &graphclient.GroupMembershipOrder{}
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.GroupMembershipOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.GroupMembershipOrderField(*cmd.OrderBy),
		}
	}

	// filter options
	id := cmd.Config.String("group-id")
	if id == "" {
		o, err := client.GetAllGroupMemberships(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.GroupMembershipOrder{order})
		cobra.CheckErr(err)
		return consoleOutput(o)
	}

	where := graphclient.GroupMembershipWhereInput{
		GroupID: &id,
	}

	o, err := client.GetGroupMemberships(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &where, []*graphclient.GroupMembershipOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
