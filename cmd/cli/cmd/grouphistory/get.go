package grouphistory

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing groupHistory",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(getCmd)
	getCmd.Flags().StringP("id", "i", "", "id to query")
}

// get an existing groupHistory in the platform
func get(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	// filter options
	id := cmd.Config.String("id")
	if id != "" {
		o, err := client.GetGroupHistories(ctx, &openlaneclient.GroupHistoryWhereInput{
			Ref: &id,
		})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllGroupHistories(ctx)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
