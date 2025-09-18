//go:build cli

package trustcenterdoc

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// command represents the base trustcenterdoc command when called without any subcommands
var command = &cobra.Command{
	Use:   "trust-center-doc",
	Aliases: []string{"trustcenterdoc"},
	Short: "the subcommands for working with trust center documents",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the trust center docs in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the trust center docs and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllTrustCenterDocs:
		var nodes []*openlaneclient.GetAllTrustCenterDocs_TrustCenterDocs_Edges_Node

		for _, i := range v.TrustCenterDocs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenterDocs:
		var nodes []*openlaneclient.GetTrustCenterDocs_TrustCenterDocs_Edges_Node

		for _, i := range v.TrustCenterDocs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenterDocByID:
		e = v.TrustCenterDoc
	case *openlaneclient.CreateTrustCenterDoc:
		e = v.CreateTrustCenterDoc.TrustCenterDoc
	case *openlaneclient.UpdateTrustCenterDoc:
		e = v.UpdateTrustCenterDoc.TrustCenterDoc
	case *openlaneclient.DeleteTrustCenterDoc:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.GetAllTrustCenterDocs_TrustCenterDocs_Edges_Node

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.GetAllTrustCenterDocs_TrustCenterDocs_Edges_Node
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
func tableOutput(out []openlaneclient.GetAllTrustCenterDocs_TrustCenterDocs_Edges_Node) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Title", "Category", "TrustCenterID", "FileID", "Visibility", "Tags", "CreatedAt", "UpdatedAt")
	for _, i := range out {
		trustCenterID := ""
		if i.TrustCenterID != nil {
			trustCenterID = *i.TrustCenterID
		}

		fileID := ""
		if i.FileID != nil {
			fileID = *i.FileID
		}

		visibility := ""
		if i.Visibility != nil {
			visibility = string(*i.Visibility)
		}

		createdAt := ""
		if i.CreatedAt != nil {
			createdAt = i.CreatedAt.Format("2006-01-02 15:04:05")
		}

		updatedAt := ""
		if i.UpdatedAt != nil {
			updatedAt = i.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		writer.AddRow(i.ID, i.Title, i.Category, trustCenterID, fileID, visibility, strings.Join(i.Tags, ","), createdAt, updatedAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteTrustCenterDoc) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTrustCenterDoc.DeletedID)

	writer.Render()
}
