package control

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing control",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "control id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("ref-code", "r", "", "the unique reference code of the control")
	updateCmd.Flags().StringP("description", "d", "", "description of the control")

	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the control")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the control")
	updateCmd.Flags().StringSliceP("add-editors", "e", []string{}, "group ID(s) given editor access to the control")
	updateCmd.Flags().StringSliceP("add-viewers", "w", []string{}, "group ID(s) given viewer access to the control")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateControlInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("control id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	refCode := cmd.Config.String("ref-code")
	if refCode != "" {
		input.RefCode = &refCode
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	addPrograms := cmd.Config.Strings("add-programs")
	if len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	removePrograms := cmd.Config.Strings("remove-programs")
	if len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

	addEditors := cmd.Config.Strings("add-editors")
	if len(addEditors) > 0 {
		input.AddEditorIDs = addEditors
	}

	addViewers := cmd.Config.Strings("add-viewers")
	if len(addViewers) > 0 {
		input.AddViewerIDs = addViewers
	}

	return id, input, nil
}

// update an existing control in the platform
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

	o, err := client.UpdateControl(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
