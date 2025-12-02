//go:build cli

package mappedcontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	openlaneclient "github.com/theopenlane/go-client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new mapped control",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("relation", "r", "", "description of how the two controls are related")
	createCmd.Flags().Int64P("confidence", "c", 0, "percentage (0-100) of confidence in the mapping")
	createCmd.Flags().StringP("mapping-type", "m", "equal", "the type of mapping between the two controls, e.g. subset, intersect, equal, superset")
	createCmd.Flags().StringP("source", "s", enums.MappingSourceManual.String(), "the source of the mapping, e.g. manual, suggested, imported")

	createCmd.Flags().StringSlice("from-control-ids", nil, "the ids of the controls to map from")
	createCmd.Flags().StringSlice("to-control-ids", nil, "the ids of the controls to map to")
	createCmd.Flags().StringSlice("from-subcontrol-ids", nil, "the ids of the subcontrols to map from")
	createCmd.Flags().StringSlice("to-subcontrol-ids", nil, "the ids of the subcontrols to map to")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateMappedControlInput, err error) {
	input.FromControlIDs = cmd.Config.Strings("from-control-ids")
	input.ToControlIDs = cmd.Config.Strings("to-control-ids")
	input.FromSubcontrolIDs = cmd.Config.Strings("from-subcontrol-ids")
	input.ToSubcontrolIDs = cmd.Config.Strings("to-subcontrol-ids")

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

	input.Source = enums.ToMappingSource(cmd.Config.String("source"))

	return input, nil
}

// create a new mappedControl
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

	o, err := client.CreateMappedControl(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
