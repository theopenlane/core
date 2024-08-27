package entitlement

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing entitlement",
	Long: `The entitlement update command updates an existing entitlement in the platform.
The command requires the entitlement id associated with the entitlement and the fields to update.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "entitlement id to update")
	updateCmd.Flags().DurationP("expires_at", "e", 0, "expiration time of the entitlement")
	updateCmd.Flags().Bool("cancelled", false, "whether the entitlement is cancelled")
	updateCmd.Flags().StringP("external_customer_id", "c", "", "external customer id")
	updateCmd.Flags().StringP("external_subscription_id", "s", "", "external subscription id")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateEntitlementInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("entitlement id")
	}

	cancelled := cmd.Config.Bool("cancelled")
	externalCustomerID := cmd.Config.String("external_customer_id")
	externalSubscriptionID := cmd.Config.String("external_subscription_id")
	expiresAt := cmd.Config.Duration("expires_at")

	input.Cancelled = &cancelled

	if externalCustomerID != "" {
		input.ExternalCustomerID = &externalCustomerID
	}

	if externalSubscriptionID != "" {
		input.ExternalSubscriptionID = &externalSubscriptionID
	}

	if expiresAt != 0 {
		input.ExpiresAt = lo.ToPtr(time.Now().Add(expiresAt))
	}

	return id, input, nil
}

// update an existing entitlement
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateEntitlement(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
