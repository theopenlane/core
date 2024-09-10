package integrationhistory

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base integrationHistory command when called without any subcommands
var command = &cobra.Command{
	Use:   "integration-history",
	Short: "the subcommands for working with integrationHistories",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the integrationHistories in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the integrationHistories and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllIntegrationHistories:
		var nodes []*openlaneclient.GetAllIntegrationHistories_IntegrationHistories_Edges_Node

		for _, i := range v.IntegrationHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetIntegrationHistories:
		var nodes []*openlaneclient.GetIntegrationHistories_IntegrationHistories_Edges_Node

		for _, i := range v.IntegrationHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.IntegrationHistory

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.IntegrationHistory
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
func tableOutput(out []openlaneclient.IntegrationHistory) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Ref", "Operation", "UpdatedAt", "UpdatedBy")
	for _, i := range out {
		writer.AddRow(i.ID, *i.Ref, i.Operation, *i.UpdatedAt, *i.UpdatedBy)
	}

	writer.Render()
}
