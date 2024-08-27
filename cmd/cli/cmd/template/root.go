package templates

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base template command when called without any subcommands
var command = &cobra.Command{
	Use:   "template",
	Short: "the subcommands for working with the templates",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the out in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the out in JSON format
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the out and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllTemplates:
		var nodes []*openlaneclient.GetAllTemplates_Templates_Edges_Node

		for _, i := range v.Templates.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetTemplateByID:
		e = v.Template
	case *openlaneclient.CreateTemplate:
		e = v.CreateTemplate.Template
	case *openlaneclient.UpdateTemplate:
		e = v.UpdateTemplate.Template
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Template

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Template
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
func tableOutput(out []openlaneclient.Template) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "JSON")
	for _, i := range out {
		// this doesn't visually show you the json in the table but leaving it in for now
		writer.AddRow(i.ID, i.Name, *i.Description, i.Jsonconfig)
	}

	writer.Render()
}
