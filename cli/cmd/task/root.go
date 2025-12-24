//go:build cli

package task

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
)

// command represents the base task command when called without any subcommands
var command = &cobra.Command{
	Use:   "task",
	Short: "the subcommands for working with tasks",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the tasks in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the tasks and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllTasks:
		var nodes []*graphclient.GetAllTasks_Tasks_Edges_Node

		if v != nil {
			for _, i := range v.Tasks.Edges {
				nodes = append(nodes, i.Node)
			}
		}

		e = nodes
	case *graphclient.GeTasks:
		var nodes []*graphclient.GeTasks_Tasks_Edges_Node

		if v != nil {
			for _, i := range v.Tasks.Edges {
				nodes = append(nodes, i.Node)
			}
		}

		e = nodes
	case *graphclient.GeTaskByID:
		e = v.Task
	case *graphclient.CreateBulkCSVTask:
		e = v.CreateBulkCSVTask.Tasks
	case *graphclient.CreateTask:
		e = v.CreateTask.Task
	case *graphclient.UpdateTask:
		e = v.UpdateTask.Task
	case *graphclient.DeleteTask:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlane.Task

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlane.Task
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
func tableOutput(out []openlane.Task) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Title", "Details", "Category", "Assignee", "Assigner", "Status", "Due")

	for _, i := range out {
		assignee := *i.AssigneeID
		if i.Assignee != nil {
			assignee = i.Assignee.DisplayName
		}

		assigner := *i.AssignerID
		if i.Assigner != nil {
			assigner = i.Assigner.DisplayName
		}

		var dueDate string
		if i.Due != nil {
			dueDate = i.Due.String()
		}

		writer.AddRow(i.ID, i.DisplayID, i.Title, *i.Details, *i.Category, assignee, assigner, i.Status, dueDate)

		if i.Comments != nil && len(i.Comments.Edges) > 0 {
			writer.AddRow("----------------------------------------")
			writer.AddRow("COMMENTS", "CREATEDBY", "CREATEDAT")
			writer.AddRow("----------------------------------------")
			for _, c := range i.Comments.Edges {
				writer.AddRow(c.Node.Text, *c.Node.CreatedBy, *c.Node.CreatedAt)
			}

			writer.AddRow("") // blank row
		}
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteTask) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTask.DeletedID)

	writer.Render()
}
