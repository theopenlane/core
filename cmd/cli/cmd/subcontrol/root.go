//go:build cli

package subcontrol

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base subcontrol command when called without any subcommands
var command = &cobra.Command{
	Use:   "subcontrol",
	Short: "the subcommands for working with subcontrols",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the subcontrols in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the subcontrols and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllSubcontrols:
		var nodes []*openlaneclient.GetAllSubcontrols_Subcontrols_Edges_Node

		for _, i := range v.Subcontrols.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetSubcontrols:
		var nodes []*openlaneclient.GetSubcontrols_Subcontrols_Edges_Node

		for _, i := range v.Subcontrols.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetSubcontrolByID:
		e = v.Subcontrol
	case *openlaneclient.CreateSubcontrol:
		e = v.CreateSubcontrol.Subcontrol
	case *openlaneclient.UpdateSubcontrol:
		e = v.UpdateSubcontrol.Subcontrol
	case *openlaneclient.DeleteSubcontrol:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Subcontrol

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Subcontrol
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
func tableOutput(out []openlaneclient.Subcontrol) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Status", "Parent Control")

	for _, i := range out {
		writer.AddRow(i.ID, i.RefCode, *i.Description, *i.Status, i.ControlID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteSubcontrol) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteSubcontrol.DeletedID)

	writer.Render()
}
