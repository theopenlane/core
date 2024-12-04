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
	createCmd.Flags().StringP("name", "n", "", "name of the subcontrol")
	createCmd.Flags().StringP("description", "d", "", "description of the subcontrol")
	createCmd.Flags().StringP("status", "s", "", "status of the subcontrol")
	createCmd.Flags().StringP("type", "t", "", "type of the subcontrol")
	createCmd.Flags().StringP("version", "v", "", "version of the subcontrol")
	createCmd.Flags().StringP("number", "u", "", "number of the subcontrol")
	createCmd.Flags().StringP("family", "f", "", "family of the subcontrol")
	createCmd.Flags().StringP("class", "l", "", "class of the subcontrol")
	createCmd.Flags().StringP("source", "o", "", "source of the subcontrol")
	createCmd.Flags().StringP("mapped-frameworks", "m", "", "mapped frameworks with the subcontrol")

	createCmd.Flags().StringP("evidence", "e", "", "evidence of the subcontrol")
	createCmd.Flags().String("implementation-status", "", "implementation status of the subcontrol")
	createCmd.Flags().StringP("implementation-verification", "r", "", "implementation verification of the subcontrol")

	createCmd.Flags().StringSliceP("controls", "c", []string{}, "control ID(s) associated with the subcontrol")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateSubcontrolInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.ControlIDs = cmd.Config.Strings("controls")
	if len(input.ControlIDs) == 0 {
		return input, cmd.NewRequiredFieldMissingError("control ID(s)")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = &status
	}

	subcontrolType := cmd.Config.String("type")
	if subcontrolType != "" {
		input.SubcontrolType = &subcontrolType
	}

	version := cmd.Config.String("version")
	if version != "" {
		input.Version = &version
	}

	number := cmd.Config.String("number")
	if number != "" {
		input.SubcontrolNumber = &number
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

	evidence := cmd.Config.String("evidence")
	if evidence != "" {
		input.ImplementationEvidence = &evidence
	}

	implementationStatus := cmd.Config.String("implementation-status")
	if implementationStatus != "" {
		input.ImplementationStatus = &implementationStatus
	}

	implementationVerification := cmd.Config.String("implementation-verification")
	if implementationVerification != "" {
		input.ImplementationVerification = &implementationVerification
	}

	return input, nil
}

// create a new subcontrol
func create(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateSubcontrol(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
