//go:build cli

package trustcentersubprocessors

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing trust center subprocessor",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)
	getCmd.Flags().StringP("id", "i", "", "trust center subprocessor id to query")
}

// get an existing trust center subprocessor in the platform
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

	// if a trust center subprocessor ID is provided, filter on that trust center subprocessor, otherwise get all
	if id != "" {
		o, err := client.GetTrustCenterSubprocessorByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.TrustCenterSubprocessorOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.TrustCenterSubprocessorOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.TrustCenterSubprocessorOrderField(*cmd.OrderBy),
		}
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllTrustCenterSubprocessors(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.TrustCenterSubprocessorOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
