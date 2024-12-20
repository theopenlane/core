package controlobjective

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing controlObjective",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "controlObjective id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the control objective")
	updateCmd.Flags().StringP("description", "d", "", "description of the control objective")
	updateCmd.Flags().StringP("status", "s", "", "status of the control objective")
	updateCmd.Flags().StringP("type", "t", "", "type of the control objective")
	updateCmd.Flags().StringP("version", "v", "", "version of the control objective")
	updateCmd.Flags().StringP("control-number", "c", "", "number of the control objective")
	updateCmd.Flags().StringP("family", "f", "", "family of the control objective")
	updateCmd.Flags().StringP("class", "l", "", "class associated with the control objective")
	updateCmd.Flags().StringP("source", "o", "", "source of the control objective")
	updateCmd.Flags().StringP("mapped-frameworks", "m", "", "mapped frameworks")
	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the risk")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the risk")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateControlObjectiveInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("controlObjective id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	addPrograms := cmd.Config.Strings("add-programs")
	if len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	removePrograms := cmd.Config.Strings("remove-programs")
	if len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

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

	return id, input, nil
}

// update an existing controlObjective in the platform
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

	o, err := client.UpdateControlObjective(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
