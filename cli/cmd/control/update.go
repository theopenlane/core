//go:build cli

package control

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
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
	updateCmd.Flags().StringP("ref-code", "r", "", "the unique reference code of the control")
	updateCmd.Flags().StringP("description", "d", "", "description of the control")
	updateCmd.Flags().StringP("source", "s", "", "source of the control, e.g. framework, template, custom, etc.")
	updateCmd.Flags().StringP("category", "c", "", "category of the control")
	updateCmd.Flags().StringP("category-id", "a", "", "category id of the control")
	updateCmd.Flags().StringP("subcategory", "b", "", "subcategory of the control")
	updateCmd.Flags().StringP("status", "t", "", "status of the control")
	updateCmd.Flags().StringP("control-type", "y", "", "type of the control e.g. preventive, detective, corrective, or deterrent")
	updateCmd.Flags().StringSliceP("mapped-categories", "m", []string{}, "mapped categories of the control to other standards")
	updateCmd.Flags().StringP("framework-id", "f", "", "framework of the control")

	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the control")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the control")
	updateCmd.Flags().StringSliceP("add-editors", "e", []string{}, "group ID(s) given editor access to the control")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateControlInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("control id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	refCode := cmd.Config.String("ref-code")
	if refCode != "" {
		input.RefCode = &refCode
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
	}

	controlType := cmd.Config.String("control-type")
	if controlType != "" {
		input.ControlType = enums.ToControlType(controlType)
	}

	mappedCategories := cmd.Config.Strings("mapped-categories")
	if len(mappedCategories) > 0 {
		input.MappedCategories = mappedCategories
	}

	frameworkID := cmd.Config.String("framework-id")
	if frameworkID != "" {
		input.StandardID = &frameworkID
	}

	addPrograms := cmd.Config.Strings("add-programs")
	if len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	removePrograms := cmd.Config.Strings("remove-programs")
	if len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

	addEditors := cmd.Config.Strings("add-editors")
	if len(addEditors) > 0 {
		input.AddEditorIDs = addEditors
	}

	return id, input, nil
}

// update an existing control in the platform
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

	o, err := client.UpdateControl(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
