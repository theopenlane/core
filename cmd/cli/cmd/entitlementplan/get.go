package entitlementplan

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get existing entitlement plans",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "entitlement plan id to query")
}

// get retrieves plans from the platform
func get(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id := cmd.Config.String("id")

	// if an plan ID is provided, filter on that plan, otherwise get all
	if id == "" {
		// get all plans, will be filtered for the authorized organization(s)
		p, err := client.GetAllEntitlementPlans(ctx)
		cobra.CheckErr(err)

		return consoleOutput(p)
	}

	p, err := client.GetEntitlementPlanByID(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(p)
}
