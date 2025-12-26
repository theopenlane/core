//go:build cli

package narrative

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new narrative",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the narrative")
	createCmd.Flags().StringP("description", "d", "", "description of the narrative")
	createCmd.Flags().StringP("details", "t", "", "details of the narrative")
	createCmd.Flags().StringP("details-file", "f", "", "file containing the details of the narrative")

	createCmd.Flags().StringSliceP("programs", "p", []string{}, "program ID(s) associated with the narrative")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateNarrativeInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.ProgramIDs = cmd.Config.Strings("programs")

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

	return input, nil
}

// create a new narrative
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

	o, err := client.CreateNarrative(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
