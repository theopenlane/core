//go:build cli

package asset

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base asset command when called without any subcommands
var command = &cobra.Command{
	Use:   "asset",
	Short: "the subcommands for working with Assets",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Assets in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Assets and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllAssets:
		var nodes []*graphclient.GetAllAssets_Assets_Edges_Node

		for _, i := range v.Assets.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetAssets:
		var nodes []*graphclient.GetAssets_Assets_Edges_Node

		for _, i := range v.Assets.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetAssetByID:
		e = v.Asset
	case *graphclient.CreateAsset:
		e = v.CreateAsset.Asset
	case *graphclient.UpdateAsset:
		e = v.UpdateAsset.Asset
	case *graphclient.DeleteAsset:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Asset

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Asset
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
func tableOutput(out []graphclient.Asset) {
	// create a table writer

	// TODO: add additional columns to the table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID")
	for _, i := range out {
		writer.AddRow(i.ID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteAsset) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteAsset.DeletedID)

	writer.Render()
}
