package internalpolicy

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing internal policy",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)
	getCmd.Flags().StringP("id", "i", "", "internal policy id to query")
}

// get an existing internal policy in the platform
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

	// if an internal policy ID is provided, filter on that internal policy, otherwise get all
	if id != "" {
		o, err := client.GetInternalPolicyByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllInternalPolicies(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
