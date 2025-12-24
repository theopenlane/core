//go:build cli

package tokens

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get personal access tokens",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "personal access token id to query")
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

	// filter options
	id := cmd.Config.String("id")

	order := &graphclient.PersonalAccessTokenOrder{}
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.PersonalAccessTokenOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.PersonalAccessTokenOrderField(*cmd.OrderBy),
		}
	}
	// if an id is provided, filter on the id, otherwise get all
	if id == "" {
		o, err := client.GetAllPersonalAccessTokens(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.PersonalAccessTokenOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	o, err := client.GetPersonalAccessTokenByID(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
