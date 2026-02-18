//go:build cli

package workflowAssignment

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base workflowAssignment command when called without any subcommands
var command = &cobra.Command{
	Use:   "workflowAssignment",
	Short: "the subcommands for working with WorkflowAssignments",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the WorkflowAssignments in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the WorkflowAssignments and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllWorkflowAssignments:
		var nodes []*graphclient.GetAllWorkflowAssignments_WorkflowAssignments_Edges_Node

		for _, i := range v.WorkflowAssignments.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetWorkflowAssignments:
		var nodes []*graphclient.GetWorkflowAssignments_WorkflowAssignments_Edges_Node

		for _, i := range v.WorkflowAssignments.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetWorkflowAssignmentByID:
		e = v.WorkflowAssignment
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.WorkflowAssignment

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.WorkflowAssignment
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
func tableOutput(out []graphclient.WorkflowAssignment) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Status", "Role", "Label", "Required", "Targets")
	for _, i := range out {
		label := ""
		if i.Label != nil {
			label = *i.Label
		}

		targetCount := 0
		if i.WorkflowAssignmentTargets != nil {
			targetCount = len(i.WorkflowAssignmentTargets.Edges)
		}

		writer.AddRow(i.ID, i.DisplayID, i.Status, i.Role, label, i.Required, targetCount)
	}

	writer.Render()
}

// consoleOutputWithTargets prints the output with target details
func consoleOutputWithTargets(e *graphclient.GetWorkflowAssignments) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	tableOutputWithTargets(e.WorkflowAssignments.Edges)

	return nil
}

// tableOutputWithTargets prints assignments with their target details
func tableOutputWithTargets(edges []*graphclient.GetWorkflowAssignments_WorkflowAssignments_Edges) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Status", "Role", "Label", "TargetType", "TargetUserID", "TargetGroupID")
	for _, edge := range edges {
		i := edge.Node
		label := ""
		if i.Label != nil {
			label = *i.Label
		}

		targets := i.WorkflowAssignmentTargets.Edges
		if len(targets) == 0 {
			writer.AddRow(i.ID, i.DisplayID, i.Status, i.Role, label, "-", "-", "-")
			continue
		}

		for idx, targetEdge := range targets {
			t := targetEdge.Node
			targetUserID := "-"
			if t.TargetUserID != nil {
				targetUserID = *t.TargetUserID
			}
			targetGroupID := "-"
			if t.TargetGroupID != nil {
				targetGroupID = *t.TargetGroupID
			}

			if idx == 0 {
				writer.AddRow(i.ID, i.DisplayID, i.Status, i.Role, label, t.TargetType, targetUserID, targetGroupID)
			} else {
				writer.AddRow("", "", "", "", "", t.TargetType, targetUserID, targetGroupID)
			}
		}
	}

	writer.Render()
}

