//go:build cli

package register

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	models "github.com/theopenlane/core/pkg/openapi"
)

var command = &cobra.Command{
	Use:   "register",
	Short: "register a new user",
	Run: func(cmd *cobra.Command, args []string) {
		err := register(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().StringP("email", "e", "", "email of the user")
	command.Flags().StringP("password", "p", "", "password of the user")
	command.Flags().StringP("first-name", "f", "", "first name of the user")
	command.Flags().StringP("last-name", "l", "", "last name of the user")
}

// validateRegister validates the required fields for the command
func validateRegister() (*models.RegisterRequest, error) {
	email := cmd.Config.String("email")
	if email == "" {
		return nil, cmd.NewRequiredFieldMissingError("email")
	}

	firstName := cmd.Config.String("first-name")
	if firstName == "" {
		return nil, cmd.NewRequiredFieldMissingError("first name")
	}

	lastName := cmd.Config.String("last-name")
	if lastName == "" {
		return nil, cmd.NewRequiredFieldMissingError("last name")
	}

	password := cmd.Config.String("password")
	if password == "" {
		return nil, cmd.NewRequiredFieldMissingError("password")
	}

	return &models.RegisterRequest{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}, nil
}

// register registers a new user in the platform
func register(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClient(ctx)
	cobra.CheckErr(err)

	var s []byte

	input, err := validateRegister()
	cobra.CheckErr(err)

	registration, err := client.Register(ctx, input)
	cobra.CheckErr(err)

	s, err = json.Marshal(registration)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}
