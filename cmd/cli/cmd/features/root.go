package feature

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/cli/tables"
)

// cmd represents the base feature command when called without any subcommands
var command = &cobra.Command{
	Use:   "feature",
	Short: "the subcommands for working with features",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetFeatures:
		var nodes []*openlaneclient.GetFeatures_Features_Edges_Node

		for _, i := range v.Features.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetFeatureByID:
		e = v.Feature
	case *openlaneclient.CreateFeature:
		e = v.CreateFeature.Feature
	case *openlaneclient.UpdateFeature:
		e = v.UpdateFeature.Feature
	case *openlaneclient.DeleteFeature:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Feature

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Feature
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
func tableOutput(out []openlaneclient.Feature) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "DisplayName", "Enabled", "Description")

	for _, i := range out {
		desc := ""
		if i.Description != nil {
			desc = *i.Description
		}

		writer.AddRow(i.ID, i.Name, *i.DisplayName, i.Enabled, desc)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteFeature) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteFeature.DeletedID)

	writer.Render()
}
