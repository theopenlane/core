package entitlementplanfeatures

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get existing plan features",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "feature id to query")
}

// get retrieves plan features from the platform
func get(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id := cmd.Config.String("id")

	// if an feature ID is provided, filter on that feature, otherwise get all
	if id == "" {
		// get all features, will be filtered for the authorized organization(s)
		out, err := client.GetAllEntitlementPlanFeatures(ctx)
		cobra.CheckErr(err)

		return consoleOutput(out)
	}

	out, err := client.GetEntitlementPlanFeatureByID(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(out)
}
