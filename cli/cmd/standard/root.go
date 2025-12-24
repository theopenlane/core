//go:build cli

package standard

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base standard command when called without any subcommands
var command = &cobra.Command{
	Use:   "standard",
	Short: "the subcommands for working with standards",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the standards in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the standards and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllStandards:
		var nodes []*graphclient.GetAllStandards_Standards_Edges_Node

		for _, i := range v.Standards.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeStandards:
		var nodes []*graphclient.GeStandards_Standards_Edges_Node

		for _, i := range v.Standards.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeStandardByID:
		e = v.Standard
	case *graphclient.CreateStandard:
		e = v.CreateStandard.Standard
	case *graphclient.UpdateStandard:
		e = v.UpdateStandard.Standard
	case *graphclient.DeleteStandard:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlane.Standard

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlane.Standard
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
func tableOutput(out []openlane.Standard) {
	// create a table writer
	// TODO: add additional columns to the table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Version", "Category", "GoverningBody", "Domain Count", "Control Count")
	for _, i := range out {
		writer.AddRow(i.ID, *i.ShortName,
			*i.Version,
			*i.StandardType,
			*i.GoverningBody,
			len(i.Domains),
			i.Controls.TotalCount)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteStandard) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteStandard.DeletedID)

	writer.Render()
}
