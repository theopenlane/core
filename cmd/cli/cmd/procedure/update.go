//go:build cli

package procedure

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/shared/enums"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing procedure",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "procedure id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the procedure")
	updateCmd.Flags().StringP("details", "d", "", "description of the procedure")
	updateCmd.Flags().StringP("status", "s", "", "status of the procedure")
	updateCmd.Flags().StringP("type", "t", "", "type of the procedure")
	updateCmd.Flags().StringP("revision", "v", "v0.1", "revision of the procedure")
	updateCmd.Flags().StringP("editor-group-id", "g", "", "editor group id")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateProcedureInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("procedure id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToDocumentStatus(status)
	}

	procedureType := cmd.Config.String("type")
	if procedureType != "" {
		input.ProcedureType = &procedureType
	}

	revision := cmd.Config.String("revision")
	if revision != "" {
		input.Revision = &revision
	}

	editorGroupID := cmd.Config.String("editor-group-id")
	if editorGroupID != "" {
		input.AddEditorIDs = []string{editorGroupID}
	}

	return id, input, nil
}

// update an existing procedure in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateProcedure(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
