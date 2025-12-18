//go:build cli

package contact

import (
	"context"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new contact",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "full name of the contact")
	createCmd.Flags().StringP("email", "e", "", "email address of the contact")
	createCmd.Flags().StringP("phone", "p", "", "phone number of the contact")
	createCmd.Flags().StringP("title", "t", "", "title of the contact")
	createCmd.Flags().StringP("company", "c", "", "company of the contact")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the contact")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateContactInput, err error) {
	// validation of required fields for the create command
	name := cmd.Config.String("name")
	if name == "" {
		return input, cmd.NewRequiredFieldMissingError("contact name")
	}

	input.FullName = lo.ToPtr(name)

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

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new contact
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateContact(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
