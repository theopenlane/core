package procedure

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new procedure",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the procedure")
	createCmd.Flags().StringP("description", "d", "", "description of the procedure")
	createCmd.Flags().StringP("status", "s", "", "status of the procedure")
	createCmd.Flags().StringP("type", "t", "", "type of the procedure")
	createCmd.Flags().StringP("version", "v", "v0.1", "version of the procedure")
	createCmd.Flags().StringP("purpose", "p", "", "purpose and scope of the procedure")
	createCmd.Flags().StringP("background", "b", "", "background information of the procedure")
	createCmd.Flags().StringP("satisfies", "S", "", "satisfies which controls")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateProcedureInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = &status
	}

	procedureType := cmd.Config.String("type")
	if procedureType != "" {
		input.ProcedureType = &procedureType
	}

	version := cmd.Config.String("version")
	if version != "" {
		input.Version = &version
	}

	purpose := cmd.Config.String("purpose")
	if purpose != "" {
		input.PurposeAndScope = &purpose
	}

	background := cmd.Config.String("background")
	if background != "" {
		input.Background = &background
	}

	satisfies := cmd.Config.String("satisfies")
	if satisfies != "" {
		input.Satisfies = &satisfies
	}

	return input, nil
}

// create a new procedure
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

	o, err := client.CreateProcedure(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
