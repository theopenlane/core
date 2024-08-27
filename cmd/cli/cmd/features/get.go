package feature

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get existing features",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "feature id to query")
}

// get an existing feature in the platform
func get(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id := cmd.Config.String("id")

	if id != "" {
		f, err := client.GetFeatureByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(f)
	}

	// get all features, will be filtered for the authorized organization(s)
	features, err := client.GetFeatures(ctx, nil)
	cobra.CheckErr(err)

	return consoleOutput(features)
}
