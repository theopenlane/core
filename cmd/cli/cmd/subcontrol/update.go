//go:build cli

package subcontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	openlaneclient "github.com/theopenlane/go-client/genclient"
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
	updateCmd.Flags().StringP("ref-code", "r", "", "the unique reference code of the subcontrol")
	updateCmd.Flags().StringP("description", "d", "", "description of the subcontrol")
	updateCmd.Flags().StringP("source", "s", "", "source of the subcontrol, e.g. framework, template, custom, etc.")
	updateCmd.Flags().StringP("category", "a", "", "category of the subcontrol")
	updateCmd.Flags().StringP("category-id", "c", "", "category id of the subcontrol")
	updateCmd.Flags().StringP("subcategory", "b", "", "subcategory of the subcontrol")
	updateCmd.Flags().StringP("status", "t", "", "status of the subcontrol")
	updateCmd.Flags().StringP("control-type", "y", "", "type of the subcontrol e.g. preventive, detective, corrective, or deterrent")
	updateCmd.Flags().StringSliceP("mapped-categories", "m", []string{}, "mapped categories of the subcontrol to other standards")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateSubcontrolInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("subcontrol id")
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

	return id, input, nil
}

// update an existing subcontrol in the platform
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

	o, err := client.UpdateSubcontrol(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
