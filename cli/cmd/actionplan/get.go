//go:build cli

package actionPlan

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing actionPlan",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "actionplan id to query")
}

// get an existing actionPlan in the platform
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

	// if an actionplan ID is provided, filter on that actionPlan, otherwise get all
	if id != "" {
		o, err := client.GetActionPlanByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.ActionPlanOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.ActionPlanOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.ActionPlanOrderField(*cmd.OrderBy),
		}
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllActionPlans(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.ActionPlanOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
