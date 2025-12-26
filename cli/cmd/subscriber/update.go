//go:build cli

package subscribers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
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
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("email", "e", "", "email address of the subscriber to update")
	updateCmd.Flags().StringP("phone-number", "p", "", "phone number to add or update on the subscriber")
}

// updateValidation validates the input flags provided by the user
func updateValidation() (email string, input graphclient.UpdateSubscriberInput, err error) {
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
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	email, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateSubscriber(ctx, email, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
