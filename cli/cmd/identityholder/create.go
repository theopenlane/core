//go:build cli

package identityHolder

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new identityHolder",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "full name of the identity holder")
	createCmd.Flags().StringP("email", "e", "", "email address of the identity holder")
	createCmd.Flags().StringP("type", "y", "", "type of identity holder (e.g. EMPLOYEE, CONTRACTOR)")
	createCmd.Flags().StringP("title", "t", "", "job title of the identity holder")
	createCmd.Flags().StringP("department", "d", "", "department of the identity holder")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateIdentityHolderInput, err error) {
	input.FullName = cmd.Config.String("name")
	if input.FullName == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.Email = cmd.Config.String("email")
	if input.Email == "" {
		return input, cmd.NewRequiredFieldMissingError("email")
	}

	identityType := cmd.Config.String("type")
	if identityType != "" {
		input.IdentityHolderType = enums.ToIdentityHolderType(identityType)
	}

	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	department := cmd.Config.String("department")
	if department != "" {
		input.Department = &department
	}

	return input, nil
}

// create a new identityHolder
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

	o, err := client.CreateIdentityHolder(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
