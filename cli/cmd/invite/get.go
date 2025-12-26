//go:build cli

package invite

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get organization invitation(s)",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "invite id to query")
}

// get an organization invitation(s)
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
	id := cmd.Config.String("id")

	if id != "" {
		o, err := client.GetInviteByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.InviteOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.InviteOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.InviteOrderField(*cmd.OrderBy),
		}
	}

	o, err := client.GetAllInvites(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.InviteOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
