//go:build cli

package invite

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	openlaneclient "github.com/theopenlane/go-client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create and invitation to join a organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("email", "e", "", "destination email for the invitation")
	createCmd.Flags().StringP("role", "r", "member", "role for the user in the organization (admin, member)")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateInviteInput, err error) {
	input.Recipient = cmd.Config.String("email")
	if input.Recipient == "" {
		return input, cmd.NewRequiredFieldMissingError("email")
	}

	input.Role = enums.ToRole(cmd.Config.String("role"))

	return input, nil
}

// create an invitation in the platform
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

	o, err := client.CreateInvite(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
