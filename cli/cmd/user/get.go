//go:build cli

package user

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/theopenlane/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get details of existing user",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "user id to query")
	getCmd.Flags().BoolP("self", "s", true, "get only details of the current user")
}

// get retrieves all users or a specific user by ID
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

	// if a user ID is provided, filter on that user, otherwise get all
	if id != "" {
		o, err := client.GetUserByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	if cmd.Config.Bool("self") {
		o, err := client.GetSelf(ctx)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	o, err := client.GetAllUsers(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, nil)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
