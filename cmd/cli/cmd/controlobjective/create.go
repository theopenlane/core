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
	createCmd.Flags().StringP("description", "d", "", "description of the control objective")
	createCmd.Flags().StringP("status", "s", "", "status of the control objective")
	createCmd.Flags().StringP("type", "t", "", "type of the control objective")
	createCmd.Flags().StringP("version", "v", "", "version of the control objective")
	createCmd.Flags().StringP("control-number", "c", "", "number of the control objective")
	createCmd.Flags().StringP("family", "f", "", "family of the control objective")
	createCmd.Flags().StringP("class", "l", "", "class associated with the control objective")
	createCmd.Flags().StringP("source", "o", "", "source of the control objective")
	createCmd.Flags().StringP("mapped-frameworks", "m", "", "mapped frameworks")
	createCmd.Flags().StringSliceP("programs", "p", []string{}, "program ID(s) associated with the control objective")
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

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = &status
	}

	controlObjectiveType := cmd.Config.String("type")
	if controlObjectiveType != "" {
		input.ControlObjectiveType = &controlObjectiveType
	}

	version := cmd.Config.String("version")
	if version != "" {
		input.Version = &version
	}

	controlNumber := cmd.Config.String("control-number")
	if controlNumber != "" {
		input.ControlNumber = &controlNumber
	}

	family := cmd.Config.String("family")
	if family != "" {
		input.Family = &family
	}

	class := cmd.Config.String("class")
	if class != "" {
		input.Class = &class
	}

	source := cmd.Config.String("source")
	if source != "" {
		input.Source = &source
	}

	mappedFrameworks := cmd.Config.String("mapped-frameworks")
	if mappedFrameworks != "" {
		input.MappedFrameworks = &mappedFrameworks
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
