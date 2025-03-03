package controlobjective

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new controlObjective",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the control objective")

	createCmd.Flags().StringSliceP("programs", "p", []string{}, "program ID(s) associated with the control objective")
	createCmd.Flags().StringSliceP("editors", "e", []string{}, "group ID(s) given editor access to the control objective")
	createCmd.Flags().StringSliceP("viewers", "w", []string{}, "group ID(s) given viewer access to the control objective")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateControlObjectiveInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.ProgramIDs = cmd.Config.Strings("programs")

	viewerGroupIDs := cmd.Config.Strings("viewers")
	if len(viewerGroupIDs) > 0 {
		input.ViewerIDs = viewerGroupIDs
	}

	editorGroupIDs := cmd.Config.Strings("editors")
	if len(editorGroupIDs) > 0 {
		input.EditorIDs = editorGroupIDs
	}

	return input, nil
}

// create a new controlObjective
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

	o, err := client.CreateControlObjective(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
