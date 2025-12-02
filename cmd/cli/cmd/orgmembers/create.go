//go:build cli

package orgmembers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "add user to a organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("org-id", "o", "", "organization id")
	createCmd.Flags().StringP("user-id", "u", "", "user id")
	createCmd.Flags().StringP("role", "r", "member", "role to assign the user (member, admin)")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateOrgMembershipInput, err error) {
	input.UserID = cmd.Config.String("user-id")
	if input.UserID == "" {
		return input, cmd.NewRequiredFieldMissingError("user id")
	}

	// role defaults to `member` so it is not required
	role := cmd.Config.String("role")

	r, err := cmd.GetRoleEnum(role)
	cobra.CheckErr(err)

	input.Role = &r

	oID := cmd.Config.String("org-id")
	if oID != "" {
		input.OrganizationID = oID
	}

	return input, nil
}

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

	o, err := client.CreateOrgMembership(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
