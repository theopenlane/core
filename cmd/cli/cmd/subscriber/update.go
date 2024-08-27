package subscribers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update subscriber details",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("email", "e", "", "email address of the subscriber to update")
	updateCmd.Flags().StringP("phone-number", "p", "", "phone number to add or update on the subscriber")
}

// updateValidation validates the input flags provided by the user
func updateValidation() (email string, input openlaneclient.UpdateSubscriberInput, err error) {
	email = cmd.Config.String("email")
	if email == "" {
		return email, input, cmd.NewRequiredFieldMissingError("email")
	}

	phone := cmd.Config.String("phone-number")

	input.PhoneNumber = &phone

	return email, input, nil
}

// update a subscriber details
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	email, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateSubscriber(ctx, email, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
