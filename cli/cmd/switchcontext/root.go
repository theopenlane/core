//go:build cli

package switchcontext

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/cli/cmd"
	models "github.com/theopenlane/core/common/openapi"
)

var command = &cobra.Command{
	Use:   "switch",
	Short: "switch organization contexts",
	Run: func(cmd *cobra.Command, args []string) {
		err := switchOrg(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().StringP("target-org", "t", "", "target organization to switch to")
}

// validate validates the required fields for the command
func validate() (*models.SwitchOrganizationRequest, error) {
	input := &models.SwitchOrganizationRequest{}

	input.TargetOrganizationID = cmd.Config.String("target-org")
	if input.TargetOrganizationID == "" {
		return nil, cmd.NewRequiredFieldMissingError("target organization")
	}

	return input, nil
}

// switchOrg switches the organization context
func switchOrg(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)

	input, err := validate()
	cobra.CheckErr(err)

	resp, err := client.Switch(ctx, input)
	cobra.CheckErr(err)

	fmt.Printf("Successfully switched to organization: %s!\n", input.TargetOrganizationID)

	// store auth tokens
	if err := cmd.StoreToken(&oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}); err != nil {
		return err
	}

	// store session cookies
	cmd.StoreSessionCookies(client)

	fmt.Println("auth tokens successfully stored in keychain")

	return nil
}
