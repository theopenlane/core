package control

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new control",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("ref-code", "r", "", "the unique reference code of the control")
	createCmd.Flags().StringP("description", "d", "", "description of the control")

	createCmd.Flags().StringSliceP("editors", "e", []string{}, "group ID(s) given editor access to the control")
	createCmd.Flags().StringSliceP("viewers", "w", []string{}, "group ID(s) given viewer access to the control")
	createCmd.Flags().StringSliceP("programs", "p", []string{}, "program ID(s) associated with the control")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateControlInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.RefCode = cmd.Config.String("ref-code")
	if input.RefCode == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.ProgramIDs = cmd.Config.Strings("programs")

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	editors := cmd.Config.Strings("editors")
	if len(editors) > 0 {
		input.EditorIDs = editors
	}

	viewers := cmd.Config.Strings("viewers")
	if len(viewers) > 0 {
		input.ViewerIDs = viewers
	}

	return input, nil
}

// create a new control
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

	o, err := client.CreateControl(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
