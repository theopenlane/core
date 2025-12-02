//go:build cli

package program

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/shared/enums"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new program",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the program")
	createCmd.Flags().StringP("description", "d", "", "description of the program")
	createCmd.Flags().Bool("auditor-ready", false, "auditor ready")
	createCmd.Flags().Bool("auditor-write-comments", false, "auditor write comments")
	createCmd.Flags().Bool("auditor-read-comments", false, "auditor read comments")
	createCmd.Flags().Duration("start-date", 0, "time until start date")
	createCmd.Flags().Duration("end-date", 0, "time until end date")
	createCmd.Flags().StringP("status", "s", enums.ProgramStatusNotStarted.String(), "status of the program")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateProgramInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	auditorReady := cmd.Config.Bool("auditor-ready")
	if auditorReady {
		input.AuditorReady = &auditorReady
	}

	auditorWriteComments := cmd.Config.Bool("auditor-write-comments")
	if auditorWriteComments {
		input.AuditorWriteComments = &auditorWriteComments
	}

	auditorReadComments := cmd.Config.Bool("auditor-read-comments")
	if auditorReadComments {
		input.AuditorReadComments = &auditorReadComments
	}

	startDate := cmd.Config.Duration("start-date")
	if startDate != 0 {
		startDate := time.Now().Add(startDate)
		input.StartDate = &startDate
	}

	endDate := cmd.Config.Duration("end-date")
	if endDate != 0 {
		endDate := time.Now().Add(endDate)
		input.EndDate = &endDate
	}

	status := cmd.Config.String("status")
	input.Status = enums.ToProgramStatus(status)

	return input, nil
}

// create a new program
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

	o, err := client.CreateProgram(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
