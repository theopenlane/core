package trustcenter

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// command represents the base trustcenter command when called without any subcommands
var command = &cobra.Command{
	Use:   "trustcenter",
	Short: "the subcommands for working with trustcenters",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the trustcenters in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the trustcenters and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllTrustCenters:
		var nodes []*openlaneclient.GetAllTrustCenters_TrustCenters_Edges_Node

		for _, i := range v.TrustCenters.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenters:
		var nodes []*openlaneclient.GetTrustCenters_TrustCenters_Edges_Node

		for _, i := range v.TrustCenters.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTrustCenterByID:
		e = v.TrustCenter
	case *openlaneclient.CreateTrustCenter:
		e = v.CreateTrustCenter.TrustCenter
	case *openlaneclient.UpdateTrustCenter:
		e = v.UpdateTrustCenter.TrustCenter
	case *openlaneclient.DeleteTrustCenter:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.GetAllTrustCenters_TrustCenters_Edges_Node

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.GetAllTrustCenters_TrustCenters_Edges_Node
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
func tableOutput(out []openlaneclient.GetAllTrustCenters_TrustCenters_Edges_Node) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Slug", "CustomDomainID", "OwnerID", "Tags", "CreatedAt", "UpdatedAt")
	for _, i := range out {
		customDomainID := ""
		if i.CustomDomainID != nil {
			customDomainID = *i.CustomDomainID
		}

		ownerID := ""
		if i.OwnerID != nil {
			ownerID = *i.OwnerID
		}

		createdAt := ""
		if i.CreatedAt != nil {
			createdAt = i.CreatedAt.Format("2006-01-02 15:04:05")
		}

		updatedAt := ""
		if i.UpdatedAt != nil {
			updatedAt = i.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		writer.AddRow(i.ID, i.Slug, customDomainID, ownerID, strings.Join(i.Tags, ","), createdAt, updatedAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteTrustCenter) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTrustCenter.DeletedID)

	writer.Render()
}
