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
	updateCmd.Flags().StringP("name", "n", "", "name of the control")
	updateCmd.Flags().StringP("description", "d", "", "description of the control")
	updateCmd.Flags().StringP("status", "s", "", "status of the control")
	updateCmd.Flags().StringP("type", "t", "", "type of the control")
	updateCmd.Flags().StringP("version", "v", "", "version of the control")
	updateCmd.Flags().StringP("number", "u", "", "number of the control")
	updateCmd.Flags().StringP("family", "f", "", "family of the control")
	updateCmd.Flags().StringP("class", "c", "", "class of the control")
	updateCmd.Flags().StringP("source", "o", "", "source of the control")
	updateCmd.Flags().StringP("mapped-frameworks", "m", "", "mapped frameworks with the control")
	updateCmd.Flags().StringP("satisfies", "a", "", "control objectives satisfied by the control")

	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the risk")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the risk")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateControlInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("control id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = &status
	}

	controlType := cmd.Config.String("type")
	if controlType != "" {
		input.ControlType = &controlType
	}

	version := cmd.Config.String("version")
	if version != "" {
		input.Version = &version
	}

	number := cmd.Config.String("number")
	if number != "" {
		input.ControlNumber = &number
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

	frameworks := cmd.Config.String("mapped-frameworks")
	if frameworks != "" {
		input.MappedFrameworks = &frameworks
	}

	satisfies := cmd.Config.String("satisfies")
	if satisfies != "" {
		input.Satisfies = &satisfies
	}

	addPrograms := cmd.Config.Strings("add-programs")
	if len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	removePrograms := cmd.Config.Strings("remove-programs")
	if len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

	return id, input, nil
}

// update an existing control in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateControl(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
