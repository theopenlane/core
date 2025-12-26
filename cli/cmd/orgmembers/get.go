//go:build cli

package orgmembers

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get existing members of a organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("org-id", "o", "", "org id to query")
}

func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	var order *graphclient.OrgMembershipOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.OrgMembershipOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.OrgMembershipOrderField(*cmd.OrderBy),
		}
	}

	where := graphclient.OrgMembershipWhereInput{}

	// filter options
	id := cmd.Config.String("id")

	if id != "" {
		where.OrganizationID = &id
	}

	o, err := client.GetOrgMemberships(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &where, []*graphclient.OrgMembershipOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
