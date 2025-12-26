//go:build cli

package groupsetting

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get group settings",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "group setting id to retrieve")
}

// get group settings from the platform
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

	var order *graphclient.GroupSettingOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.GroupSettingOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.GroupSettingOrderField(*cmd.OrderBy),
		}
	}

	// if setting ID is not provided, get settings which will automatically filter by group id
	if id == "" {
		o, err := client.GetAllGroupSettings(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.GroupSettingOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	o, err := client.GetGroupSettingByID(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
