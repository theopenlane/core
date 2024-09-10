package integrations

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base integration command when called without any subcommands
var command = &cobra.Command{
	Use:   "integration",
	Short: "the subcommands for working with integrations",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllIntegrations:
		var nodes []*openlaneclient.GetAllIntegrations_Integrations_Edges_Node

		for _, i := range v.Integrations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetIntegrationByID:
		e = v.Integration
	case *openlaneclient.CreateIntegration:
		e = v.CreateIntegration.Integration
	case *openlaneclient.UpdateIntegration:
		e = v.UpdateIntegration.Integration
	case *openlaneclient.DeleteIntegration:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Integration

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Integration
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
func tableOutput(out []openlaneclient.Integration) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Kind")
	for _, i := range out {
		writer.AddRow(i.ID, i.Name, *i.Description, i.Kind)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteIntegration) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteIntegration.DeletedID)

	writer.Render()
}
