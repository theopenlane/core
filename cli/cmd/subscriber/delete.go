//go:build cli

package subscribers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove a subscriber from an organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := delete(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("email", "e", "", "email address of the subscriber to delete")
	deleteCmd.Flags().StringP("organization-id", "o", "", "organization ID of the subscriber to delete, only required when using a personal access token for the request")
}

// deleteValidation validates the required fields for the command
func deleteValidation() (string, *string, error) {
	email := cmd.Config.String("email")
	if email == "" {
		return "", nil, cmd.NewRequiredFieldMissingError("email")
	}

	orgID := cmd.Config.String("organization-id")
	if orgID == "" {
		return email, nil, nil
	}

	return email, &orgID, nil
}

// delete an existing organization subscriber
func delete(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	email, orgID, err := deleteValidation()
	cobra.CheckErr(err)

	o, err := client.DeleteSubscriber(ctx, email, orgID)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
