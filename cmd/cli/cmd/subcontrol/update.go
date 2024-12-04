package subcontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing subcontrol",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "subcontrol id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the subcontrol")
	updateCmd.Flags().StringP("description", "d", "", "description of the subcontrol")
	updateCmd.Flags().StringP("status", "s", "", "status of the subcontrol")
	updateCmd.Flags().StringP("type", "t", "", "type of the subcontrol")
	updateCmd.Flags().StringP("version", "v", "", "version of the subcontrol")
	updateCmd.Flags().StringP("number", "u", "", "number of the subcontrol")
	updateCmd.Flags().StringP("family", "f", "", "family of the subcontrol")
	updateCmd.Flags().StringP("class", "l", "", "class of the subcontrol")
	updateCmd.Flags().StringP("source", "o", "", "source of the subcontrol")
	updateCmd.Flags().StringP("mapped-frameworks", "m", "", "mapped frameworks with the subcontrol")

	updateCmd.Flags().StringP("evidence", "e", "", "evidence of the subcontrol")
	updateCmd.Flags().String("implementation-status", "", "implementation status of the subcontrol")
	updateCmd.Flags().StringP("implementation-verification", "r", "", "implementation verification of the subcontrol")

	updateCmd.Flags().StringSlice("add-controls", []string{}, "add control(s) to the subcontrol")
	updateCmd.Flags().StringSlice("remove-controls", []string{}, "remove control(s) from the subcontrol")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateSubcontrolInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("subcontrol id")
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

	addControls := cmd.Config.Strings("add-controls")
	if len(addControls) > 0 {
		input.AddControlIDs = addControls
	}

	removeControls := cmd.Config.Strings("remove-controls")
	if len(removeControls) > 0 {
		input.RemoveControlIDs = removeControls
	}

	return id, input, nil
}

// update an existing subcontrol in the platform
func update(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateSubcontrol(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
