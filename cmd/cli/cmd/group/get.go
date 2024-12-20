package group

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing group",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "group id to query")
}

// get an existing group in the platform
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

	// if an group ID is provided, filter on that group, otherwise get all
	if id != "" {
		o, err := client.GetGroupByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all o, will be filtered for the authorized organization(s)
	o, err := client.GetAllGroups(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
