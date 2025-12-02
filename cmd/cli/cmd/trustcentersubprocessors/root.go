//go:build cli

package trustcentersubprocessors

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base trustcentersubprocessors command when called without any subcommands
var command = &cobra.Command{
	Use:   "trustcentersubprocessors",
	Short: "the subcommands for working with trust center subprocessors",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the trust center subprocessors in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the trust center subprocessors and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllTrustCenterSubprocessors:
		var nodes []*openlaneclient.GetAllTrustCenterSubprocessors_TrustCenterSubprocessors_Edges_Node

		for _, i := range v.TrustCenterSubprocessors.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenterSubprocessors:
		var nodes []*openlaneclient.GetTrustCenterSubprocessors_TrustCenterSubprocessors_Edges_Node

		for _, i := range v.TrustCenterSubprocessors.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenterSubprocessorByID:
		e = v.TrustCenterSubprocessor
	case *openlaneclient.CreateTrustCenterSubprocessor:
		e = v.CreateTrustCenterSubprocessor.TrustCenterSubprocessor
	case *openlaneclient.UpdateTrustCenterSubprocessor:
		e = v.UpdateTrustCenterSubprocessor.TrustCenterSubprocessor
	case *openlaneclient.DeleteTrustCenterSubprocessor:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.GetAllTrustCenterSubprocessors_TrustCenterSubprocessors_Edges_Node

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.GetAllTrustCenterSubprocessors_TrustCenterSubprocessors_Edges_Node
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
func tableOutput(out []openlaneclient.GetAllTrustCenterSubprocessors_TrustCenterSubprocessors_Edges_Node) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Trust Center ID", "Subprocessor Name", "Category", "Countries")
	for _, i := range out {
		trustCenterID := ""
		if i.TrustCenterID != nil {
			trustCenterID = *i.TrustCenterID
		}

		subprocessorName := i.Subprocessor.Name

		countries := strings.Join(i.Countries, ", ")

		writer.AddRow(i.ID, trustCenterID, subprocessorName, i.Category, countries)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteTrustCenterSubprocessor) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTrustCenterSubprocessor.DeletedID)

	writer.Render()
}
