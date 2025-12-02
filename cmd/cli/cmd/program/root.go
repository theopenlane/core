//go:build cli

package program

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

// command represents the base program command when called without any subcommands
var command = &cobra.Command{
	Use:   "program",
	Short: "the subcommands for working with programs",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the programs in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the programs and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllPrograms:
		var nodes []*openlaneclient.GetAllPrograms_Programs_Edges_Node

		for _, i := range v.Programs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetPrograms:
		var nodes []*openlaneclient.GetPrograms_Programs_Edges_Node

		for _, i := range v.Programs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetProgramByID:
		e = v.Program
	case *openlaneclient.CreateProgram:
		e = v.CreateProgram.Program
	case *openlaneclient.UpdateProgram:
		e = v.UpdateProgram.Program
	case *openlaneclient.DeleteProgram:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Program

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Program
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
func tableOutput(out []openlaneclient.Program) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Description", "Status", "AuditorReady", "AuditorWriteComments", "AuditorReadComments", "StartDate", "EndDate")
	for _, i := range out {
		writer.AddRow(i.ID, i.DisplayID, i.Name, *i.Description, i.Status, i.AuditorReady, i.AuditorWriteComments, i.AuditorReadComments, i.StartDate, i.EndDate)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteProgram) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteProgram.DeletedID)

	writer.Render()
}
