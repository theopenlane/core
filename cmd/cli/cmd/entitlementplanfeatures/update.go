package entitlementplanfeatures

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing entitlement plan feature",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "feature id to update")

	// fields for the plan feature metadata
	updateCmd.Flags().StringP("limit-type", "t", "", "limit type for the plan feature, e.g. requests, storage, etc.")
	updateCmd.Flags().Int64P("limit", "l", 0, "limit value for the plan feature")

	updateCmd.Flags().StringSlice("tags", []string{}, "tags associated with the plan")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateEntitlementPlanFeatureInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("plan feature id")
	}

	limitType := cmd.Config.String("limit-type")
	limit := cmd.Config.Int64("limit")

	if limitType == "" && limit > 0 {
		return id, input, cmd.NewRequiredFieldMissingError("limit type")
	}

	if limitType != "" && limit == 0 {
		return id, input, cmd.NewRequiredFieldMissingError("limit")
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

	return id, input, nil
}

// update an existing plan feature in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	f, err := client.UpdateEntitlementPlanFeature(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(f)
}
