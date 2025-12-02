//go:build cli

package narrative

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing narrative",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "narrative id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the narrative")
	updateCmd.Flags().StringP("description", "d", "", "description of the narrative")
	updateCmd.Flags().StringP("details", "t", "", "details of the narrative")
	updateCmd.Flags().StringP("details-file", "f", "", "file containing the details of the narrative")

	updateCmd.Flags().StringSlice("add-programs", []string{}, "add program(s) to the narrative")
	updateCmd.Flags().StringSlice("remove-programs", []string{}, "remove program(s) from the narrative")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input openlaneclient.UpdateNarrativeInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("narrative id")
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

	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	} else {
		detailsFile := cmd.Config.String("details-file")
		if detailsFile != "" {
			contents, err := os.ReadFile(detailsFile)
			cobra.CheckErr(err)

			details := string(contents)

			input.Details = &details
		}
	}

	addPrograms := cmd.Config.Strings("add-programs")
	if len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	removePrograms := cmd.Config.Strings("remove-programs")
	if len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

	return id, input, nil
}

// update an existing narrative in the platform
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

	o, err := client.UpdateNarrative(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
