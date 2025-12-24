//go:build cli

package standard

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing standard",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)
	getCmd.Flags().StringP("id", "i", "", "standard id to query")

}

// get an existing standard in the platform
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

	// if an standard ID is provided, filter on that standard, otherwise get all
	if id != "" {
		o, err := client.GetStandardByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	order := &graphclient.StandardOrder{}
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.StandardOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.StandardOrderField(*cmd.OrderBy),
		}
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllStandards(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.StandardOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
