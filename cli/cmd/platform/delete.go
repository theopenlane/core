//go:build cli

package platform

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete an existing platform",
	Run: func(cmd *cobra.Command, args []string) {
		err := delete(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("id", "i", "", "platform id to delete")
}

// deleteValidation validates the required fields for the command
func deleteValidation() (string, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", cmd.NewRequiredFieldMissingError("platform id")
	}

	return id, nil
}

// delete an existing platform in the platform
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

	o, err := client.DeletePlatform(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
