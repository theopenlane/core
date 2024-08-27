package apitokens

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get or list api tokens",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "api token id to query")
}

// get retrieves api tokens from the platform
func get(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	// filter options
	tokenID := cmd.Config.String("id")

	// if an api token ID is provided, filter on that api token, otherwise get all
	if tokenID == "" {
		tokens, err := client.GetAllAPITokens(ctx)
		cobra.CheckErr(err)

		return consoleOutput(tokens)
	}

	token, err := client.GetAPITokenByID(ctx, tokenID)
	cobra.CheckErr(err)

	return consoleOutput(token)
}
