//go:build cli

package notificationpreference

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing notification preference",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "notification preference id to query")
}

// get an existing notification preference
func get(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id := cmd.Config.String("id")

	if id != "" {
		o, err := client.GetNotificationPreferenceByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.NotificationPreferenceOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.NotificationPreferenceOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.NotificationPreferenceOrderField(*cmd.OrderBy),
		}
	}

	o, err := client.GetAllNotificationPreferences(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.NotificationPreferenceOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
