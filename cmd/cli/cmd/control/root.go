package control

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base control command when called without any subcommands
var command = &cobra.Command{
	Use:   "control",
	Short: "the subcommands for working with controls",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the controls in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the controls and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllControls:
		var nodes []*openlaneclient.GetAllControls_Controls_Edges_Node

		for _, i := range v.Controls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetControls:
		var nodes []*openlaneclient.GetControls_Controls_Edges_Node

		for _, i := range v.Controls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetControlByID:
		e = v.Control
	case *openlaneclient.CreateControl:
		e = v.CreateControl.Control
	case *openlaneclient.UpdateControl:
		e = v.UpdateControl.Control
	case *openlaneclient.DeleteControl:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Control

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Control
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
func tableOutput(out []openlaneclient.Control) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Source", "Category", "CategoryID", "Subcategory", "Status", "ControlType", "MappedCategories", "Standard", "Programs")

	for _, i := range out {
		programs := []string{}

		for _, p := range i.Programs.Edges {
			programs = append(programs, p.Node.Name)
		}

		stdName := ""
		if i.Standard != nil {
			if i.Standard.ShortName != nil {
				stdName = *i.Standard.ShortName
			} else {
				stdName = i.Standard.Name
			}
		}

		writer.AddRow(i.ID, i.RefCode, *i.Description, i.Source, *i.Category, *i.CategoryID, *i.Subcategory, *i.Status, i.ControlType, strings.Join(i.MappedCategories, ", "), stdName, strings.Join(programs, ", "))
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteControl) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteControl.DeletedID)

	writer.Render()
}
