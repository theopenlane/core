//go:build cli

package directorygroup

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing directory group",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "directory group id to query")
}

// get an existing directory group
func get(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id := cmd.Config.String("id")

	if id != "" {
		o, err := client.GetDirectoryGroupByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.DirectoryGroupOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.DirectoryGroupOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.DirectoryGroupOrderField(*cmd.OrderBy),
		}
	}

	o, err := client.GetAllDirectoryGroups(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.DirectoryGroupOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
