//go:build cli

package orgsetting

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get organization settings",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "organization setting id to retrieve")
}

// get an existing organization setting in the platform
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

	var order *graphclient.OrganizationSettingOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.OrganizationSettingOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.OrganizationSettingOrderField(*cmd.OrderBy),
		}
	}

	// if setting ID is not provided, get settings which will automatically filter by org id
	if id == "" {
		o, err := client.GetAllOrganizationSettings(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.OrganizationSettingOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	o, err := client.GetOrganizationSettingByID(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
