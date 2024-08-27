package entitlement

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new entitlement",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("organization_id", "o", "", "organization id associated with the entitlement")
	createCmd.Flags().StringP("plan_id", "p", "", "plan_id associated with the entitlement")
	createCmd.Flags().DurationP("expires_at", "e", 0, "expiration time of the entitlement")
	createCmd.Flags().Bool("cancelled", false, "whether the entitlement is cancelled")
	createCmd.Flags().StringP("external_customer_id", "c", "", "external customer id")
	createCmd.Flags().StringP("external_subscription_id", "s", "", "external subscription id")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the entitlement")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateEntitlementInput, err error) {
	orgID := cmd.Config.String("organization_id")
	if orgID == "" {
		return input, cmd.NewRequiredFieldMissingError("organization id")
	}

	planID := cmd.Config.String("plan_id")
	if planID == "" {
		return input, cmd.NewRequiredFieldMissingError("plan id")
	}

	cancelled := cmd.Config.Bool("cancelled")
	externalCustomerID := cmd.Config.String("external_customer_id")
	externalSubscriptionID := cmd.Config.String("external_subscription_id")
	expiresAt := cmd.Config.Duration("expires_at")
	tags := cmd.Config.Strings("tags")

	input = openlaneclient.CreateEntitlementInput{
		OrganizationID: orgID,
		PlanID:         planID,
		Cancelled:      &cancelled,
	}

	if externalCustomerID != "" {
		input.ExternalCustomerID = &externalCustomerID
	}

	if externalSubscriptionID != "" {
		input.ExternalSubscriptionID = &externalSubscriptionID
	}

	if expiresAt != 0 {
		input.ExpiresAt = lo.ToPtr(time.Now().Add(expiresAt))
	}

	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new entitlement in the platform
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateEntitlement(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
