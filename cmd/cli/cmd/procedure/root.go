//go:build cli

package procedure

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

// command represents the base procedure command when called without any subcommands
var command = &cobra.Command{
	Use:   "procedure",
	Short: "the subcommands for working with procedures",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the procedures in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the procedures and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllProcedures:
		var nodes []*openlaneclient.GetAllProcedures_Procedures_Edges_Node

		for _, i := range v.Procedures.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetProcedures:
		var nodes []*openlaneclient.GetProcedures_Procedures_Edges_Node

		for _, i := range v.Procedures.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetProcedureByID:
		e = v.Procedure
	case *openlaneclient.CreateProcedure:
		e = v.CreateProcedure.Procedure
	case *openlaneclient.CreateUploadProcedure:
		e = v.CreateUploadProcedure.Procedure
	case *openlaneclient.UpdateProcedure:
		e = v.UpdateProcedure.Procedure
	case *openlaneclient.DeleteProcedure:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Procedure

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Procedure
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
func tableOutput(out []openlaneclient.Procedure) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Details", "Status", "Type", "Revision", "ReviewDue", "ReviewFrequency", "ApprovalRequired")
	for _, i := range out {
		writer.AddRow(i.ID, i.DisplayID, i.Name, *i.Details, *i.Status, *i.ProcedureType, *i.Revision, *i.ReviewDue, *i.ReviewFrequency, *i.ApprovalRequired)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteProcedure) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteProcedure.DeletedID)

	writer.Render()
}
