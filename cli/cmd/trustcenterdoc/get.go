//go:build cli

package trustcenterdoc

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing trust center document",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "trust center document id to query")
}

// get an existing trust center document in the platform
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

	// if a trust center document ID is provided, filter on that trust center document, otherwise get all
	if id != "" {
		o, err := client.GetTrustCenterDocByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllTrustCenterDocs(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
