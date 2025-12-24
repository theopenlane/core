//go:build cli

package documentdata

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete an existing document data",
	Run: func(cmd *cobra.Command, args []string) {
		err := delete(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("id", "i", "", "document data id to delete")
}

// deleteValidation validates the required fields for the command
func deleteValidation() (string, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", cmd.NewRequiredFieldMissingError("document data id")
	}

	return id, nil
}

// delete an existing document data
func delete(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, err := deleteValidation()
	cobra.CheckErr(err)

	o, err := client.DeleteDocumentData(ctx, id)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
