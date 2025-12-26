//go:build cli

package usersetting

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/theopenlane/core/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get user settings",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "user setting id to retrieve")
}

// get retrieves user settings from the platform
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

	// if setting ID is not provided, get settings which will automatically filter by user id
	if id != "" {
		o, err := client.GetUserSettingByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	o, err := client.GetAllUserSettings(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, nil)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
