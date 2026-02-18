//go:build cli

package workflowInstance

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base workflowInstance command when called without any subcommands
var command = &cobra.Command{
	Use:   "workflowInstance",
	Short: "the subcommands for working with WorkflowInstances",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the WorkflowInstances in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the WorkflowInstances and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllWorkflowInstances:
		var nodes []*graphclient.GetAllWorkflowInstances_WorkflowInstances_Edges_Node

		for _, i := range v.WorkflowInstances.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetWorkflowInstances:
		var nodes []*graphclient.GetWorkflowInstances_WorkflowInstances_Edges_Node

		for _, i := range v.WorkflowInstances.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetWorkflowInstanceByID:
		e = v.WorkflowInstance
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.WorkflowInstance

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.WorkflowInstance
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
func tableOutput(out []graphclient.WorkflowInstance) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "State", "DefinitionID", "Assignments", "CreatedAt")
	for _, i := range out {
		createdAt := ""
		if i.CreatedAt != nil {
			createdAt = i.CreatedAt.Format("2006-01-02 15:04")
		}

		assignmentCount := 0
		if i.WorkflowAssignments != nil {
			assignmentCount = len(i.WorkflowAssignments.Edges)
		}

		writer.AddRow(i.ID, i.DisplayID, i.State, i.WorkflowDefinitionID, assignmentCount, createdAt)
	}

	writer.Render()
}

// consoleOutputWithAssignments prints a single workflow instance with assignment details
func consoleOutputWithAssignments(e *graphclient.GetWorkflowInstanceByID) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	tableOutputInstanceWithAssignments(e.WorkflowInstance.ID, e.WorkflowInstance.DisplayID, e.WorkflowInstance.State, e.WorkflowInstance.WorkflowAssignments.Edges)

	return nil
}

// consoleOutputWithAssignmentsAll prints all workflow instances with assignment details
func consoleOutputWithAssignmentsAll(e *graphclient.GetAllWorkflowInstances) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	writer := tables.NewTableWriter(command.OutOrStdout(), "InstanceID", "DisplayID", "State", "AssignmentRole", "AssignmentStatus", "TargetType", "TargetUserID", "TargetGroupID")
	for _, edge := range e.WorkflowInstances.Edges {
		inst := edge.Node
		assignments := inst.WorkflowAssignments.Edges
		if len(assignments) == 0 {
			writer.AddRow(inst.ID, inst.DisplayID, inst.State, "-", "-", "-", "-", "-")
			continue
		}

		firstRow := true
		for _, assignmentEdge := range assignments {
			a := assignmentEdge.Node
			targets := a.WorkflowAssignmentTargets.Edges
			if len(targets) == 0 {
				if firstRow {
					writer.AddRow(inst.ID, inst.DisplayID, inst.State, a.Role, a.Status, "-", "-", "-")
					firstRow = false
				} else {
					writer.AddRow("", "", "", a.Role, a.Status, "-", "-", "-")
				}
				continue
			}

			for _, targetEdge := range targets {
				t := targetEdge.Node
				targetUserID := "-"
				if t.TargetUserID != nil {
					targetUserID = *t.TargetUserID
				}
				targetGroupID := "-"
				if t.TargetGroupID != nil {
					targetGroupID = *t.TargetGroupID
				}

				if firstRow {
					writer.AddRow(inst.ID, inst.DisplayID, inst.State, a.Role, a.Status, t.TargetType, targetUserID, targetGroupID)
					firstRow = false
				} else {
					writer.AddRow("", "", "", a.Role, a.Status, t.TargetType, targetUserID, targetGroupID)
				}
			}
		}
	}

	writer.Render()

	return nil
}

// tableOutputInstanceWithAssignments prints a single instance with its assignments
func tableOutputInstanceWithAssignments(instanceID, displayID string, state any, assignments []*graphclient.GetWorkflowInstanceByID_WorkflowInstance_WorkflowAssignments_Edges) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "InstanceID", "DisplayID", "State", "AssignmentRole", "AssignmentStatus", "TargetType", "TargetUserID", "TargetGroupID")
	if len(assignments) == 0 {
		writer.AddRow(instanceID, displayID, state, "-", "-", "-", "-", "-")
		writer.Render()

		return
	}

	firstRow := true
	for _, assignmentEdge := range assignments {
		a := assignmentEdge.Node
		targets := a.WorkflowAssignmentTargets.Edges
		if len(targets) == 0 {
			if firstRow {
				writer.AddRow(instanceID, displayID, state, a.Role, a.Status, "-", "-", "-")
				firstRow = false
			} else {
				writer.AddRow("", "", "", a.Role, a.Status, "-", "-", "-")
			}
			continue
		}

		for _, targetEdge := range targets {
			t := targetEdge.Node
			targetUserID := "-"
			if t.TargetUserID != nil {
				targetUserID = *t.TargetUserID
			}
			targetGroupID := "-"
			if t.TargetGroupID != nil {
				targetGroupID = *t.TargetGroupID
			}

			if firstRow {
				writer.AddRow(instanceID, displayID, state, a.Role, a.Status, t.TargetType, targetUserID, targetGroupID)
				firstRow = false
			} else {
				writer.AddRow("", "", "", a.Role, a.Status, t.TargetType, targetUserID, targetGroupID)
			}
		}
	}

	writer.Render()
}

