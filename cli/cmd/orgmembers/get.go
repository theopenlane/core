//go:build cli

package orgmembers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
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

	where := openlane.OrgMembershipWhereInput{}

	// filter options
	id := cmd.Config.String("id")

	if id != "" {
		where.OrganizationID = &id
	}

	o, err := client.GetOrgMemberships(ctx, cmd.First, cmd.Last, &where)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
