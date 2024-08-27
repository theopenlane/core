package user

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing user",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "user id to update")
	updateCmd.Flags().StringP("first-name", "f", "", "first name of the user")
	updateCmd.Flags().StringP("last-name", "l", "", "last name of the user")
	updateCmd.Flags().StringP("display-name", "d", "", "display name of the user")
	updateCmd.Flags().StringP("email", "e", "", "email of the user")
}

// updateValidation validates the input flags provided by the user
func updateValidation() (id string, input openlaneclient.UpdateUserInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("user id")
	}

	firstName := cmd.Config.String("first-name")
	if firstName != "" {
		input.FirstName = &firstName
	}

	lastName := cmd.Config.String("last-name")
	if lastName != "" {
		input.LastName = &lastName
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	email := cmd.Config.String("email")
	if email != "" {
		input.Email = &email
	}

	// TODO: allow updates to user settings
	return id, input, nil
}

// update an existing user
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateUser(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
