//go:build cli

package subscribers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get existing subscribers of a organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().BoolP("active", "a", true, "filter on active subscribers")
	getCmd.Flags().StringP("email", "e", "", "email address of the subscriber to get")
}

// get an existing subscriber(s) of an organization
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	email := cmd.Config.String("email")

	if email != "" {
		o, err := client.GetSubscriberByEmail(ctx, email)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// filter options
	where := openlaneclient.SubscriberWhereInput{}

	active := cmd.Config.Bool("active")
	if active {
		where.Active = &active
	}

	o, err := client.GetSubscribers(ctx, nil, nil, nil, nil, &where, nil)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
