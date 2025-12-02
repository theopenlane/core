//go:build cli

package contact

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing contact",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "contact id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "full name of the contact")
	updateCmd.Flags().StringP("email", "e", "", "email address of the contact")
	updateCmd.Flags().StringP("phone", "p", "", "phone number of the contact")
	updateCmd.Flags().StringP("title", "t", "", "title of the contact")
	updateCmd.Flags().StringP("company", "c", "", "company of the contact")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateContactInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("contact id")
	}

	// validation of required fields for the update command
	name := cmd.Config.String("name")
	if name != "" {
		input.FullName = &name
	}

	email := cmd.Config.String("email")
	if email != "" {
		input.Email = &email
	}

	phone := cmd.Config.String("phone")
	if phone != "" {
		input.PhoneNumber = &phone
	}

	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	company := cmd.Config.String("company")
	if company != "" {
		input.Company = &company
	}

	return id, input, nil
}

// update an existing contact in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateContact(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
