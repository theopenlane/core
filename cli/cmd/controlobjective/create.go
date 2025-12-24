//go:build cli

package controlobjective

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new controlObjective",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the control objective")
	createCmd.Flags().StringP("desired-outcome", "o", "", "desired outcome of the control objective")
	createCmd.Flags().StringP("source", "s", "", "source of the control objective, e.g. framework, template, custom, etc.")
	createCmd.Flags().StringP("revision", "r", "", "revision of the control objective")
	createCmd.Flags().StringP("status", "t", "", "status of the control objective")
	createCmd.Flags().StringP("control-objective-type", "y", "", "type of the control objective e.g. compliance, operational, strategic, or reporting")

	createCmd.Flags().StringSliceP("controls", "c", []string{}, "control ID(s) associated with the control objective")
	createCmd.Flags().StringSliceP("programs", "p", []string{}, "program ID(s) associated with the control objective")
	createCmd.Flags().StringSliceP("editors", "e", []string{}, "group ID(s) given editor access to the control objective")
	createCmd.Flags().StringSliceP("viewers", "w", []string{}, "group ID(s) given viewer access to the control objective")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateControlObjectiveInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
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

	controlIDs := cmd.Config.Strings("controls")
	if len(controlIDs) > 0 {
		input.ControlIDs = controlIDs
	}

	programIDs := cmd.Config.Strings("programs")
	if len(programIDs) > 0 {
		input.ProgramIDs = programIDs
	}

	viewerGroupIDs := cmd.Config.Strings("viewers")
	if len(viewerGroupIDs) > 0 {
		input.ViewerIDs = viewerGroupIDs
	}

	editorGroupIDs := cmd.Config.Strings("editors")
	if len(editorGroupIDs) > 0 {
		input.EditorIDs = editorGroupIDs
	}

	return input, nil
}

// create a new controlObjective
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

	o, err := client.CreateControlObjective(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
