package subcontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new subcontrol",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("ref-code", "n", "", "the unique reference code of the subcontrol")
	createCmd.Flags().StringP("description", "d", "", "description of the subcontrol")

	createCmd.Flags().StringP("control", "c", "", "[required] control ID associated with the subcontrol")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateSubcontrolInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.RefCode = cmd.Config.String("ref-code")
	if input.RefCode == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.ControlID = cmd.Config.String("controls")
	if len(input.ControlID) == 0 {
		return input, cmd.NewRequiredFieldMissingError("control ID")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	return input, nil
}

// create a new subcontrol
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

	o, err := client.CreateSubcontrol(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
