//go:build cli

package scheduledjob

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
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
	case *graphclient.GetAllScheduledJobs:
		var nodes []*graphclient.GetAllScheduledJobs_ScheduledJobs_Edges_Node

		for _, i := range v.ScheduledJobs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetScheduledJobs:
		var nodes []*graphclient.GetScheduledJobs_ScheduledJobs_Edges_Node

		for _, i := range v.ScheduledJobs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetScheduledJobByID:
		e = v.ScheduledJob
	case *graphclient.CreateScheduledJob:
		e = v.CreateScheduledJob.ScheduledJob
	case *graphclient.UpdateScheduledJob:
		e = v.UpdateScheduledJob.ScheduledJob
	case *graphclient.DeleteScheduledJob:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.ScheduledJob

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.ScheduledJob
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
func tableOutput(out []graphclient.ScheduledJob) {
	// create a table writer
	// TODO: add additional columns to the table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID")
	for _, i := range out {
		writer.AddRow(i.ID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteScheduledJob) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteScheduledJob.DeletedID)

	writer.Render()
}
