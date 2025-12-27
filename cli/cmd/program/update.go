//go:build cli

package program

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing program",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "program id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the program")
	updateCmd.Flags().StringP("description", "d", "", "description of the program")
	updateCmd.Flags().Bool("auditor-ready", false, "auditor ready")
	updateCmd.Flags().Bool("auditor-write-comments", false, "auditor write comments")
	updateCmd.Flags().Bool("auditor-read-comments", false, "auditor read comments")
	updateCmd.Flags().Duration("start-date", 0, "time until start date")
	updateCmd.Flags().Duration("end-date", 0, "time until end date")
	updateCmd.Flags().StringP("status", "s", enums.ProgramStatusNotStarted.String(), "status of the program")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateProgramInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("program id")
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
	if status != "" {
		input.Status = enums.ToProgramStatus(status)
	}

	return id, input, nil
}

// update an existing program in the platform
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

	o, err := client.UpdateProgram(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
