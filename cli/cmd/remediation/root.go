//go:build cli

package remediation

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base remediation command when called without any subcommands
var command = &cobra.Command{
	Use:   "remediation",
	Short: "the subcommands for working with Remediations",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Remediations in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Remediations and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllRemediations:
		var nodes []*graphclient.GetAllRemediations_Remediations_Edges_Node

		for _, i := range v.Remediations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetRemediations:
		var nodes []*graphclient.GetRemediations_Remediations_Edges_Node

		for _, i := range v.Remediations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetRemediationByID:
		e = v.Remediation
	case *graphclient.CreateRemediation:
		e = v.CreateRemediation.Remediation
	case *graphclient.UpdateRemediation:
		e = v.UpdateRemediation.Remediation
	case *graphclient.DeleteRemediation:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Remediation

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Remediation
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
func tableOutput(out []graphclient.Remediation) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Intent", "DueAt", "CompletedAt")
	for _, i := range out {
		intent := ""
		if i.Intent != nil {
			intent = *i.Intent
		}

		dueAt := ""
		if i.DueAt != nil {
			dueAt = i.DueAt.String()
		}

		completedAt := ""
		if i.CompletedAt != nil {
			completedAt = i.CompletedAt.String()
		}

		writer.AddRow(i.ID, intent, dueAt, completedAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteRemediation) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteRemediation.DeletedID)

	writer.Render()
}
