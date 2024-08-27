package user

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new user",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("email", "e", "", "email of the user")
	createCmd.Flags().StringP("password", "p", "", "password of the user")
	createCmd.Flags().StringP("first-name", "f", "", "first name of the user")
	createCmd.Flags().StringP("last-name", "l", "", "last name of the user")
	createCmd.Flags().StringP("display-name", "d", "", "first name of the user")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateUserInput, err error) {
	input.Email = cmd.Config.String("email")
	if input.Email == "" {
		return input, cmd.NewRequiredFieldMissingError("email")
	}

	firstName := cmd.Config.String("first-name")
	if firstName == "" {
		return input, cmd.NewRequiredFieldMissingError("first name")
	}

	input.FirstName = &firstName

	lastName := cmd.Config.String("last-name")
	if lastName == "" {
		return input, cmd.NewRequiredFieldMissingError("last name")
	}

	input.LastName = &lastName

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = displayName
	}

	password := cmd.Config.String("password")
	if password != "" {
		input.Password = &password
	}

	return input, nil
}

// create a new user
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateUser(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
