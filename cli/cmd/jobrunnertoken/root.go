//go:build cli

package jobrunnertoken

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base jobrunnertoken command when called without any subcommands
var command = &cobra.Command{
	Use:   "jobrunnertoken",
	Short: "the subcommands for working with JobRunnerTokens",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the JobRunnerTokens in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the JobRunnerTokens and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllJobRunnerTokens:
		var nodes []*graphclient.GetAllJobRunnerTokens_JobRunnerTokens_Edges_Node

		for _, i := range v.JobRunnerTokens.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetJobRunnerTokens:
		var nodes []*graphclient.GetJobRunnerTokens_JobRunnerTokens_Edges_Node

		for _, i := range v.JobRunnerTokens.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetJobRunnerTokenByID:
		e = v.JobRunnerToken
	case *graphclient.CreateJobRunnerToken:
		e = v.CreateJobRunnerToken.JobRunnerToken
	case *graphclient.DeleteJobRunnerToken:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.JobRunnerToken

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.JobRunnerToken
		err = json.Unmarshal(s, &in)
		cobra.CheckErr(err)

		list = append(list, in)
	}

	tableOutput(list)

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the output in a table format
func tableOutput(out []graphclient.JobRunnerToken) {
	// create a table writer
	// TODO: add additional columns to the table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID")
	for _, i := range out {
		writer.AddRow(i.ID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteJobRunnerToken) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteJobRunnerToken.DeletedID)

	writer.Render()
}
