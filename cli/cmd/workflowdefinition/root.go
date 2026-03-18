//go:build cli

package workflowdefinition

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base workflowdefinition command when called without any subcommands
var command = &cobra.Command{
	Use:     "workflowdefinition",
	Aliases: []string{"workflow", "workflows", "workflow-definition", "workflow-def"},
	Short:   "the subcommands for working with workflow definitions",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllWorkflowDefinitions:
		var nodes []*graphclient.GetAllWorkflowDefinitions_WorkflowDefinitions_Edges_Node

		for _, i := range v.WorkflowDefinitions.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetWorkflowDefinitions:
		var nodes []*graphclient.GetWorkflowDefinitions_WorkflowDefinitions_Edges_Node

		for _, i := range v.WorkflowDefinitions.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetWorkflowDefinitionByID:
		e = v.WorkflowDefinition
	case *graphclient.CreateWorkflowDefinition:
		e = v.CreateWorkflowDefinition.WorkflowDefinition
	case *graphclient.UpdateWorkflowDefinition:
		e = v.UpdateWorkflowDefinition.WorkflowDefinition
	case *graphclient.DeleteWorkflowDefinition:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.WorkflowDefinition

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.WorkflowDefinition
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
func tableOutput(out []graphclient.WorkflowDefinition) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "SchemaType", "Kind", "Active", "Draft", "ApprovalMode", "UpdatedAt")

	for _, i := range out {
		approvalMode := "AUTO_SUBMIT"
		if i.DefinitionJSON != nil && i.DefinitionJSON.ApprovalSubmissionMode != "" {
			approvalMode = i.DefinitionJSON.ApprovalSubmissionMode.String()
		}

		updated := "-"
		if i.UpdatedAt != nil {
			updated = i.UpdatedAt.Format(time.RFC3339)
		}

		writer.AddRow(i.ID, i.Name, i.SchemaType, i.WorkflowKind.String(), i.Active, i.Draft, approvalMode, updated)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteWorkflowDefinition) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteWorkflowDefinition.DeletedID)

	writer.Render()
}
