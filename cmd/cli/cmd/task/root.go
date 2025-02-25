package task

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	case *openlaneclient.GetAllTasks:
		var nodes []*openlaneclient.GetAllTasks_Tasks_Edges_Node

		if v != nil {
			for _, i := range v.Tasks.Edges {
				nodes = append(nodes, i.Node)
			}
		}

		e = nodes
	case *openlaneclient.GetTasks:
		var nodes []*openlaneclient.GetTasks_Tasks_Edges_Node

		if v != nil {
			for _, i := range v.Tasks.Edges {
				nodes = append(nodes, i.Node)
			}
		}

		e = nodes
	case *openlaneclient.GetTaskByID:
		e = v.Task
	case *openlaneclient.CreateTask:
		e = v.CreateTask.Task
	case *openlaneclient.UpdateTask:
		e = v.UpdateTask.Task
	case *openlaneclient.DeleteTask:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Task

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Task
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
func tableOutput(out []openlaneclient.Task) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Title", "Description", "Details", "Category", "Assignee", "Assigner", "Status", "Due")

	for _, i := range out {
		assignee := ""
		if i.Assignee != nil {
			assignee = i.Assignee.ID
		}

		var dueDate string
		if i.Due != nil {
			dueDate = i.Due.Format(time.RFC3339)
		}

		writer.AddRow(i.ID, i.DisplayID, i.Title, *i.Description, *i.Details, *i.Category, assignee, i.Assigner.ID, i.Status, dueDate)

		if i.Comments != nil {
			writer.AddRow("----------------------------------------")
			writer.AddRow("COMMENTS", "CREATEDBY", "CREATEDAT")
			writer.AddRow("----------------------------------------")
			for _, c := range i.Comments {
				writer.AddRow(c.Text, *c.CreatedBy, *c.CreatedAt)
			}
		}

		writer.AddRow("") // blank row
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteTask) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTask.DeletedID)

	writer.Render()
}
