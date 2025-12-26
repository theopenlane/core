//go:build cli

package subcontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/go-client/graphclient"
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
	createCmd.Flags().StringP("ref-code", "r", "", "the unique reference code of the subcontrol")
	createCmd.Flags().StringP("description", "d", "", "description of the subcontrol")
	createCmd.Flags().StringP("source", "s", "", "source of the subcontrol, e.g. framework, template, custom, etc.")
	createCmd.Flags().StringP("category", "a", "", "category of the subcontrol")
	createCmd.Flags().StringP("category-id", "i", "", "category id of the subcontrol")
	createCmd.Flags().StringP("subcategory", "b", "", "subcategory of the subcontrol")
	createCmd.Flags().StringP("status", "t", "", "status of the subcontrol")
	createCmd.Flags().StringP("control-type", "y", "", "type of the subcontrol e.g. preventive, detective, corrective, or deterrent")
	createCmd.Flags().StringSliceP("mapped-categories", "m", []string{}, "mapped categories of the subcontrol to other standards")

	createCmd.Flags().StringP("control", "c", "", "[required] control ID associated with the subcontrol")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateSubcontrolInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.RefCode = cmd.Config.String("ref-code")
	if input.RefCode == "" {
		return input, cmd.NewRequiredFieldMissingError("ref-code")
	}

	input.ControlID = cmd.Config.String("control")
	if len(input.ControlID) == 0 {
		return input, cmd.NewRequiredFieldMissingError("control ID")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	source := cmd.Config.String("source")
	if source != "" {
		input.Source = enums.ToControlSource(source)
	}

	category := cmd.Config.String("category")
	if category != "" {
		input.Category = &category
	}

	categoryID := cmd.Config.String("category-id")
	if categoryID != "" {
		input.CategoryID = &categoryID
	}

	subcategory := cmd.Config.String("subcategory")
	if subcategory != "" {
		input.Subcategory = &subcategory
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToControlStatus(status)
	} else {
		input.Status = &enums.ControlStatusNotImplemented
	}

	controlType := cmd.Config.String("control-type")
	if controlType != "" {
		input.ControlType = enums.ToControlType(controlType)
	}

	mappedCategories := cmd.Config.Strings("mapped-categories")
	if len(mappedCategories) > 0 {
		input.MappedCategories = mappedCategories
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
