//go:build cli

package mappedcontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/shared/enums"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing mapped control",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "mapped control id to update")

	updateCmd.Flags().StringP("relation", "r", "", "description of how the two controls are related")
	updateCmd.Flags().Int64P("confidence", "c", 0, "percentage (0-100) of confidence in the mapping")
	updateCmd.Flags().StringP("mapping-type", "m", "", "the type of mapping between the two controls, e.g. subset, intersect, equal, superset")
	updateCmd.Flags().StringP("source", "s", "", "the source of the mapping, e.g. manual, suggested, imported")

	updateCmd.Flags().StringSlice("add-from-control-ids", nil, "the ids of the controls to map from")
	updateCmd.Flags().StringSlice("add-to-control-ids", nil, "the ids of the controls to map to")
	updateCmd.Flags().StringSlice("remove-from-control-ids", nil, "the ids of the controls to remove from the mapping")
	updateCmd.Flags().StringSlice("remove-to-control-ids", nil, "the ids of the controls to remove from the mapping")

	updateCmd.Flags().StringSlice("add-from-subcontrol-ids", nil, "the ids of the subcontrols to map from")
	updateCmd.Flags().StringSlice("add-to-subcontrol-ids", nil, "the ids of the subcontrols to map to")
	updateCmd.Flags().StringSlice("remove-from-subcontrol-ids", nil, "the ids of the subcontrols to remove from the mapping")
	updateCmd.Flags().StringSlice("remove-to-subcontrol-ids", nil, "the ids of the subcontrols to remove from the mapping")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateMappedControlInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("mapped control id")
	}

	confidence := cmd.Config.Int64("confidence")
	if confidence > 0 {
		input.Confidence = &confidence
	}

	relation := cmd.Config.String("relation")
	if relation != "" {
		input.Relation = &relation
	}

	mappingType := cmd.Config.String("mapping-type")
	if mappingType != "" {
		input.MappingType = enums.ToMappingType(mappingType)
	}

	source := cmd.Config.String("source")
	if source != "" {
		input.Source = enums.ToMappingSource(source)
	}

	addFromControlIDs := cmd.Config.Strings("add-from-control-ids")
	if len(addFromControlIDs) > 0 {
		input.AddFromControlIDs = addFromControlIDs
	}

	addToControlIDs := cmd.Config.Strings("add-to-control-ids")
	if len(addToControlIDs) > 0 {
		input.AddToControlIDs = addToControlIDs
	}

	removeFromControlIDs := cmd.Config.Strings("remove-from-control-ids")
	if len(removeFromControlIDs) > 0 {
		input.RemoveFromControlIDs = removeFromControlIDs
	}

	removeToControlIDs := cmd.Config.Strings("remove-to-control-ids")
	if len(removeToControlIDs) > 0 {
		input.RemoveToControlIDs = removeToControlIDs
	}

	addFromSubcontrolIDs := cmd.Config.Strings("add-from-subcontrol-ids")
	if len(addFromSubcontrolIDs) > 0 {
		input.AddFromSubcontrolIDs = addFromSubcontrolIDs
	}

	addToSubcontrolIDs := cmd.Config.Strings("add-to-subcontrol-ids")
	if len(addToSubcontrolIDs) > 0 {
		input.AddToSubcontrolIDs = addToSubcontrolIDs
	}

	removeFromSubcontrolIDs := cmd.Config.Strings("remove-from-subcontrol-ids")
	if len(removeFromSubcontrolIDs) > 0 {
		input.RemoveFromSubcontrolIDs = removeFromSubcontrolIDs
	}

	removeToSubcontrolIDs := cmd.Config.Strings("remove-to-subcontrol-ids")
	if len(removeToSubcontrolIDs) > 0 {
		input.RemoveToSubcontrolIDs = removeToSubcontrolIDs
	}

	return id, input, nil
}

// update an existing mappedControl in the platform
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

	o, err := client.UpdateMappedControl(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
