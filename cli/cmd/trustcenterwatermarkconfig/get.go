//go:build cli

package trustcenterwatermarkconfig

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing trust center watermark config",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "trust center watermark config id to query")
	getCmd.Flags().StringP("trust-center-id", "t", "", "trust center id to query")
}

// get an existing trust center watermark config in the platform
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

	// if a trust center watermark config ID is provided, filter on that config, otherwise get all
	if id != "" {
		o, err := client.GetTrustCenterWatermarkConfigByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		o, err := client.GetTrustCenterWatermarkConfigs(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &graphclient.TrustCenterWatermarkConfigWhereInput{
			TrustCenterID: &trustCenterID,
		}, nil)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllTrustCenterWatermarkConfigs(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, nil)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
