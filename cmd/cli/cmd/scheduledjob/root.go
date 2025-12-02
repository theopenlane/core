//go:build cli

package scheduledjob

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base scheduledjob command when called without any subcommands
var command = &cobra.Command{
	Use:   "scheduledjob",
	Short: "the subcommands for working with ScheduledJobs",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the ScheduledJobs in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the ScheduledJobs and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllScheduledJobs:
		var nodes []*openlaneclient.GetAllScheduledJobs_ScheduledJobs_Edges_Node

		for _, i := range v.ScheduledJobs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetScheduledJobs:
		var nodes []*openlaneclient.GetScheduledJobs_ScheduledJobs_Edges_Node

		for _, i := range v.ScheduledJobs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetScheduledJobByID:
		e = v.ScheduledJob
	case *openlaneclient.CreateScheduledJob:
		e = v.CreateScheduledJob.ScheduledJob
	case *openlaneclient.UpdateScheduledJob:
		e = v.UpdateScheduledJob.ScheduledJob
	case *openlaneclient.DeleteScheduledJob:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.ScheduledJob

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.ScheduledJob
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
func tableOutput(out []openlaneclient.ScheduledJob) {
	// create a table writer
	// TODO: add additional columns to the table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID")
	for _, i := range out {
		writer.AddRow(i.ID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteScheduledJob) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteScheduledJob.DeletedID)

	writer.Render()
}
