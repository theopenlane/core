//go:build cli

package mappedcontrol

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base mappedControl command when called without any subcommands
var command = &cobra.Command{
	Use:   "mapped-control",
	Short: "the subcommands for working with mapped controls",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the mappedControls in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the mappedControls and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllMappedControls:
		var nodes []*openlaneclient.GetAllMappedControls_MappedControls_Edges_Node

		for _, i := range v.MappedControls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetMappedControls:
		var nodes []*openlaneclient.GetMappedControls_MappedControls_Edges_Node

		for _, i := range v.MappedControls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetMappedControlByID:
		e = v.MappedControl
	case *openlaneclient.CreateMappedControl:
		e = v.CreateMappedControl.MappedControl
	case *openlaneclient.UpdateMappedControl:
		e = v.UpdateMappedControl.MappedControl
	case *openlaneclient.DeleteMappedControl:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.MappedControl

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.MappedControl
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
func tableOutput(out []openlaneclient.MappedControl) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "ToControls", "FromControls", "Relation", "Confidence", "MappingType", "Source")
	for _, i := range out {
		to := []string{}

		for _, j := range i.ToControls.Edges {
			to = append(to, j.Node.RefCode)
		}

		for _, j := range i.ToSubcontrols.Edges {
			to = append(to, j.Node.RefCode)
		}

		from := []string{}

		for _, j := range i.FromControls.Edges {
			from = append(from, j.Node.RefCode)
		}

		for _, j := range i.FromSubcontrols.Edges {
			from = append(from, j.Node.RefCode)
		}

		confidence := "-"
		if i.Confidence != nil {
			confidence = fmt.Sprintf("%d%%", *i.Confidence)
		}

		writer.AddRow(i.ID, strings.Join(to, ", "), strings.Join(from, ", "), *i.Relation, confidence, i.MappingType, *i.Source)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteMappedControl) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteMappedControl.DeletedID)

	writer.Render()
}
