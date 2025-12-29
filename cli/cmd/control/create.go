//go:build cli

package control

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
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
	createCmd.Flags().StringP("source", "s", "", "source of the control, e.g. framework, template, custom, etc.")
	createCmd.Flags().StringP("category", "a", "", "category of the control")
	createCmd.Flags().StringP("category-id", "i", "", "category id of the control")
	createCmd.Flags().StringP("subcategory", "b", "", "subcategory of the control")
	createCmd.Flags().StringP("status", "t", "", "status of the control")
	createCmd.Flags().StringP("control-type", "y", "", "type of the control e.g. preventive, detective, corrective, or deterrent")
	createCmd.Flags().StringSliceP("mapped-categories", "m", []string{}, "mapped categories of the control to other standards")
	createCmd.Flags().StringP("framework-id", "f", "", "framework of the control")

	createCmd.Flags().StringSliceP("editors", "e", []string{}, "group ID(s) given editor access to the control")
	createCmd.Flags().StringSliceP("programs", "p", []string{}, "program ID(s) associated with the control")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateControlInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.RefCode = cmd.Config.String("ref-code")
	if input.RefCode == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
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

	input.Status = enums.ToControlStatus(cmd.Config.String("status"))

	fmt.Println("Status:", input.Status.String())

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

	programs := cmd.Config.Strings("programs")
	if len(programs) > 0 {
		input.ProgramIDs = programs
	}

	editors := cmd.Config.Strings("editors")
	if len(editors) > 0 {
		input.EditorIDs = editors
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
