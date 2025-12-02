//go:build cli

package jobresult

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base jobresult command when called without any subcommands
var command = &cobra.Command{
	Use:   "jobresult",
	Short: "the subcommands for working with JobResults",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the JobResults in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the JobResults and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllJobResults:
		var nodes []*openlaneclient.GetAllJobResults_JobResults_Edges_Node

		for _, i := range v.JobResults.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetJobResults:
		var nodes []*openlaneclient.GetJobResults_JobResults_Edges_Node

		for _, i := range v.JobResults.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetJobResultByID:
		e = v.JobResult
	case *openlaneclient.CreateJobResult:
		e = v.CreateJobResult.JobResult
	case *openlaneclient.UpdateJobResult:
		e = v.UpdateJobResult.JobResult
	case *openlaneclient.DeleteJobResult:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.JobResult

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.JobResult
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
func tableOutput(out []openlaneclient.JobResult) {
	// create a table writer
	// TODO: add additional columns to the table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID")
	for _, i := range out {
		writer.AddRow(i.ID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteJobResult) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteJobResult.DeletedID)

	writer.Render()
}
