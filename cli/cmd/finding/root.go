//go:build cli

package finding

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base finding command when called without any subcommands
var command = &cobra.Command{
	Use:   "finding",
	Short: "the subcommands for working with Findings",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Findings in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Findings and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllFindings:
		var nodes []*graphclient.GetAllFindings_Findings_Edges_Node

		for _, i := range v.Findings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetFindings:
		var nodes []*graphclient.GetFindings_Findings_Edges_Node

		for _, i := range v.Findings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetFindingByID:
		e = v.Finding
	case *graphclient.CreateFinding:
		e = v.CreateFinding.Finding
	case *graphclient.UpdateFinding:
		e = v.UpdateFinding.Finding
	case *graphclient.DeleteFinding:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Finding

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Finding
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
func tableOutput(out []graphclient.Finding) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayName", "Severity", "Category", "Open", "Remediations", "Vulnerabilities")
	for _, i := range out {
		displayName := ""
		if i.DisplayName != nil {
			displayName = *i.DisplayName
		}

		severity := ""
		if i.Severity != nil {
			severity = string(*i.Severity)
		}

		category := ""
		if i.Category != nil {
			category = *i.Category
		}

		open := false
		if i.Open != nil {
			open = *i.Open
		}

		remediationCount := 0
		if i.Remediations != nil {
			remediationCount = len(i.Remediations.Edges)
		}

		vulnerabilityCount := 0
		if i.Vulnerabilities != nil {
			vulnerabilityCount = len(i.Vulnerabilities.Edges)
		}

		writer.AddRow(i.ID, displayName, severity, category, open, remediationCount, vulnerabilityCount)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteFinding) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteFinding.DeletedID)

	writer.Render()
}
