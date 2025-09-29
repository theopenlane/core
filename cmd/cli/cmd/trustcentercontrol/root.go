//go:build cli

package trustcentercontrol

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// command represents the base trustcentercontrol command when called without any subcommands
var command = &cobra.Command{
	Use:     "trust-center-control",
	Aliases: []string{"trustcentercontrol"},
	Short:   "the subcommands for working with trust center controls",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the trust center controls in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the trust center controls and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllTrustCenterControls:
		var nodes []*openlaneclient.GetAllTrustCenterControls_TrustCenterControls_Edges_Node

		for _, i := range v.TrustCenterControls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenterControls:
		var nodes []*openlaneclient.GetTrustCenterControls_TrustCenterControls_Edges_Node

		for _, i := range v.TrustCenterControls.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenterControlByID:
		e = v.TrustCenterControl
	case *openlaneclient.CreateTrustCenterControl:
		e = v.CreateTrustCenterControl.TrustCenterControl
	case *openlaneclient.UpdateTrustCenterControl:
		e = v.UpdateTrustCenterControl.TrustCenterControl
	case *openlaneclient.DeleteTrustCenterControl:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.GetAllTrustCenterControls_TrustCenterControls_Edges_Node

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.GetAllTrustCenterControls_TrustCenterControls_Edges_Node
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
func tableOutput(out []openlaneclient.GetAllTrustCenterControls_TrustCenterControls_Edges_Node) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "ControlID", "TrustCenterID", "Tags", "CreatedAt", "UpdatedAt")

	for _, i := range out {
		trustCenterID := ""
		if i.TrustCenterID != nil {
			trustCenterID = *i.TrustCenterID
		}

		createdAt := ""
		if i.CreatedAt != nil {
			createdAt = i.CreatedAt.Format("2006-01-02 15:04:05")
		}

		updatedAt := ""
		if i.UpdatedAt != nil {
			updatedAt = i.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		writer.AddRow(i.ID, i.ControlID, trustCenterID, strings.Join(i.Tags, ","), createdAt, updatedAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteTrustCenterControl) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTrustCenterControl.DeletedID)

	writer.Render()
}
