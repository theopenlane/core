//go:build cli

package scan

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base scan command when called without any subcommands
var command = &cobra.Command{
	Use:   "scan",
	Short: "the subcommands for working with Scans",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Scans in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Scans and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllScans:
		var nodes []*graphclient.GetAllScans_Scans_Edges_Node

		for _, i := range v.Scans.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetScans:
		var nodes []*graphclient.GetScans_Scans_Edges_Node

		for _, i := range v.Scans.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetScanByID:
		e = v.Scan
	case *graphclient.CreateScan:
		e = v.CreateScan.Scan
	case *graphclient.UpdateScan:
		e = v.UpdateScan.Scan
	case *graphclient.DeleteScan:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Scan

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Scan
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
func tableOutput(out []graphclient.Scan) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Target", "Type", "Status", "CreatedAt")
	for _, i := range out {
		createdAt := ""
		if i.CreatedAt != nil {
			createdAt = i.CreatedAt.Format("2006-01-02 15:04")
		}

		writer.AddRow(i.ID, i.Target, i.ScanType, i.Status, createdAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteScan) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteScan.DeletedID)

	writer.Render()
}
