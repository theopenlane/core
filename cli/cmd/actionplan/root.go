//go:build cli

package actionPlan

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base actionPlan command when called without any subcommands
var command = &cobra.Command{
	Use:   "actionPlan",
	Short: "the subcommands for working with ActionPlans",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the ActionPlans in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the ActionPlans and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllActionPlans:
		var nodes []*graphclient.GetAllActionPlans_ActionPlans_Edges_Node

		for _, i := range v.ActionPlans.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetActionPlans:
		var nodes []*graphclient.GetActionPlans_ActionPlans_Edges_Node

		for _, i := range v.ActionPlans.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetActionPlanByID:
		e = v.ActionPlan
	case *graphclient.CreateActionPlan:
		e = v.CreateActionPlan.ActionPlan
	case *graphclient.UpdateActionPlan:
		e = v.UpdateActionPlan.ActionPlan
	case *graphclient.DeleteActionPlan:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.ActionPlan

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.ActionPlan
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
func tableOutput(out []graphclient.ActionPlan) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Status", "Priority", "DueDate")
	for _, i := range out {
		status := ""
		if i.Status != nil {
			status = string(*i.Status)
		}

		priority := ""
		if i.Priority != nil {
			priority = string(*i.Priority)
		}

		dueDate := ""
		if i.DueDate != nil {
			dueDate = i.DueDate.Format("2006-01-02")
		}

		writer.AddRow(i.ID, i.Name, status, priority, dueDate)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteActionPlan) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteActionPlan.DeletedID)

	writer.Render()
}
