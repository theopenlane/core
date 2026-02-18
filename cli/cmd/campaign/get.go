//go:build cli

package campaign

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing campaign",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "campaign id to query")
	getCmd.Flags().BoolP("show-targets", "t", false, "show campaign target details")
}

// get an existing campaign in the platform
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
	showTargets := cmd.Config.Bool("show-targets")

	// if an campaign ID is provided, filter on that campaign, otherwise get all
	if id != "" {
		o, err := client.GetCampaignByID(ctx, id)
		cobra.CheckErr(err)

		if showTargets {
			return consoleOutputWithTargets(o)
		}

		return consoleOutput(o)
	}

	var order *graphclient.CampaignOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.CampaignOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.CampaignOrderField(*cmd.OrderBy),
		}
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllCampaigns(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.CampaignOrder{order})
	cobra.CheckErr(err)

	if showTargets {
		return consoleOutputWithTargetsAll(o)
	}

	return consoleOutput(o)
}
