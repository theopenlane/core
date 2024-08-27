package entitlementplanfeatures

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new entitlement plan feature",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(createCmd)

	createCmd.Flags().StringP("plan-id", "p", "", "plan id to use for the plan feature")
	createCmd.Flags().StringP("feature-id", "f", "", "feature id to use for the plan feature")

	// fields for the plan feature metadata
	createCmd.Flags().StringP("limit-type", "t", "", "limit type for the plan feature, e.g. requests, storage, etc.")
	createCmd.Flags().Int64P("limit", "l", 0, "limit value for the plan feature")

	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the plan")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateEntitlementPlanFeatureInput, err error) {
	planID := cmd.Config.String("plan-id")
	if planID == "" {
		return input, cmd.NewRequiredFieldMissingError("plan id")
	}

	featureID := cmd.Config.String("feature-id")
	if featureID == "" {
		return input, cmd.NewRequiredFieldMissingError("feature id")
	}

	input = openlaneclient.CreateEntitlementPlanFeatureInput{
		PlanID:    planID,
		FeatureID: featureID,
	}

	limitType := cmd.Config.String("limit-type")
	limit := cmd.Config.Int64("limit")

	if limitType == "" && limit > 0 {
		return input, cmd.NewRequiredFieldMissingError("limit type")
	}

	if limitType != "" && limit == 0 {
		return input, cmd.NewRequiredFieldMissingError("limit")
	}

	if limitType != "" && limit > 0 {
		input.Metadata = map[string]interface{}{
			"limit_type": limitType,
			"limit":      limit,
		}
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new plan feature in the platform
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	f, err := client.CreateEntitlementPlanFeature(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(f)
}
