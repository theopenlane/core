//go:build cli

package jobrunnerregistrationtoken

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base jobrunnerregistrationtoken command when called without any subcommands
var command = &cobra.Command{
	Use:   "jobrunnerregistrationtoken",
	Short: "the subcommands for working with JobRunnerRegistrationTokens",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the JobRunnerRegistrationTokens in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the JobRunnerRegistrationTokens and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.CreateJobRunnerRegistrationToken:
		e = v.CreateJobRunnerRegistrationToken.JobRunnerRegistrationToken
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var in openlaneclient.JobRunnerRegistrationToken
	err = json.Unmarshal(s, &in)
	cobra.CheckErr(err)

	tableOutput([]openlaneclient.JobRunnerRegistrationToken{in})

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the output in a table format
func tableOutput(out []openlaneclient.JobRunnerRegistrationToken) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Token", "ExpiresAt")
	for _, i := range out {
		writer.AddRow(i.ID, i.Token, i.ExpiresAt)
	}

	writer.Render()
}
