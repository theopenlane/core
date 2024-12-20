package invite

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete an invitation to join an organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := delete(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("id", "i", "", "id of the invitation to delete")
}

// deleteValidation validates the required fields for the command
func deleteValidation() (string, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", cmd.NewRequiredFieldMissingError("id")
	}

	return id, nil
}

// delete an invitation to join an organization
func delete(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, err := deleteValidation()
	cobra.CheckErr(err)

	o, err := client.DeleteInvite(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
