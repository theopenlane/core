//go:build cli

package controlobjective

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/go-client/graphclient"
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
	updateCmd.Flags().StringP("desired-outcome", "o", "", "desired outcome of the control objective")
	updateCmd.Flags().StringP("source", "s", "", "source of the control objective, e.g. framework, template, custom, etc.")
	updateCmd.Flags().StringP("revision", "v", "", "revision of the control objective")
	updateCmd.Flags().StringP("status", "t", "", "status of the control objective")
	updateCmd.Flags().StringP("control-objective-type", "y", "", "type of the control objective e.g. compliance, operational, strategic, or reporting")

	updateCmd.Flags().StringSlice("add-controls", []string{}, "add control(s) to the control objective")
	updateCmd.Flags().StringSlice("remove-controls", []string{}, "remove control(s) from the control objective")
	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the control objective")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the control objective")
	updateCmd.Flags().StringSliceP("add-editors", "e", []string{}, "group ID(s) given editor access to the control objective")
	updateCmd.Flags().StringSliceP("add-viewers", "w", []string{}, "group ID(s) given viewer access to the control objective")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateControlObjectiveInput, err error) {
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

	desiredOutcome := cmd.Config.String("desired-outcome")
	if desiredOutcome != "" {
		input.DesiredOutcome = &desiredOutcome
	}

	source := cmd.Config.String("source")
	if source != "" {
		input.Source = enums.ToControlSource(source)
	}

	revision := cmd.Config.String("revision")
	if revision != "" {
		input.Revision = &revision
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToObjectiveStatus(status)
	}

	controlObjectiveType := cmd.Config.String("control-objective-type")
	if controlObjectiveType != "" {
		input.ControlObjectiveType = &controlObjectiveType
	}

	addControls := cmd.Config.Strings("add-controls")
	if len(addControls) > 0 {
		input.AddControlIDs = addControls
	}

	removeControls := cmd.Config.Strings("remove-controls")
	if len(removeControls) > 0 {
		input.RemoveControlIDs = removeControls
	}

	addPrograms := cmd.Config.Strings("add-programs")
	if len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	removePrograms := cmd.Config.Strings("remove-programs")
	if len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

	viewerGroupIDs := cmd.Config.Strings("add-viewers")
	if len(viewerGroupIDs) > 0 {
		input.AddViewerIDs = viewerGroupIDs
	}

	editorGroupIDs := cmd.Config.Strings("add-editors")
	if len(editorGroupIDs) > 0 {
		input.AddEditorIDs = editorGroupIDs
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
