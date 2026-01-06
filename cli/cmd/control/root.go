//go:build cli

package control

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
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
	case *graphclient.GetAllControls:
		var nodes []*graphclient.GetAllControls_Controls_Edges_Node

		for _, i := range v.Controls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetControls:
		var nodes []*graphclient.GetControls_Controls_Edges_Node

		for _, i := range v.Controls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetControlByID:
		e = v.Control
	case *graphclient.CreateControl:
		e = v.CreateControl.Control
	case *graphclient.UpdateControl:
		e = v.UpdateControl.Control
	case *graphclient.DeleteControl:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Control

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Control
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
func tableOutput(out []graphclient.Control) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Source", "Category", "CategoryID", "Subcategory", "Status", "MappedCategories", "Standard")

	for _, i := range out {
		refFramework := "-"
		if i.ReferenceFramework != nil {
			refFramework = *i.ReferenceFramework
		}

		cat := "-"
		if i.Category != nil {
			cat = *i.Category
		}

		catID := "-"
		if i.CategoryID != nil {
			catID = *i.CategoryID
		}

		subcat := "-"
		if i.Subcategory != nil {
			subcat = *i.Subcategory
		}

		writer.AddRow(i.ID, i.RefCode, *i.Description, i.Source, cat, catID, subcat, i.Status.String(), strings.Join(i.MappedCategories, ", "), refFramework)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteControl) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteControl.DeletedID)

	writer.Render()
}
