//go:build cli

package identityHolder

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing identityHolder",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "identityholder id to update")
	updateCmd.Flags().StringP("full-name", "n", "", "full name of the identity holder")
	updateCmd.Flags().StringP("email", "e", "", "email address of the identity holder")
	updateCmd.Flags().StringP("alternate-email", "", "", "alternate email address")
	updateCmd.Flags().StringP("phone-number", "p", "", "phone number")
	updateCmd.Flags().StringP("environment", "", "", "environment of the identity holder")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateIdentityHolderInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("identityholder id")
	}

	fullName := cmd.Config.String("full-name")
	if fullName != "" {
		input.FullName = &fullName
	}

	email := cmd.Config.String("email")
	if email != "" {
		input.Email = &email
	}

	alternateEmail := cmd.Config.String("alternate-email")
	if alternateEmail != "" {
		input.AlternateEmail = &alternateEmail
	}

	phoneNumber := cmd.Config.String("phone-number")
	if phoneNumber != "" {
		input.PhoneNumber = &phoneNumber
	}

	environment := cmd.Config.String("environment")
	if environment != "" {
		input.EnvironmentName = &environment
	}

	return id, input, nil
}

// update an existing identityHolder in the platform
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

	o, err := client.UpdateIdentityHolder(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
