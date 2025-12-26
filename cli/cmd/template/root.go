//go:build cli

package templates

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
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
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the out and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllTemplates:
		var nodes []*graphclient.GetAllTemplates_Templates_Edges_Node

		for _, i := range v.Templates.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetTemplateByID:
		e = v.Template
	case *graphclient.CreateTemplate:
		e = v.CreateTemplate.Template
	case *graphclient.UpdateTemplate:
		e = v.UpdateTemplate.Template
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Template

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Template
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
func tableOutput(out []graphclient.Template) {
	for _, i := range out {
		writer := tables.NewTableWriter(command.OutOrStdout())

		writer.SetHeaders("ID", "Name", "Description")
		writer.AddRow(i.ID, i.Name, *i.Description)
		writer.Render()

		tmp, err := json.Marshal(i.Jsonconfig)
		cobra.CheckErr(err)

		var res map[string]interface{}
		err = json.Unmarshal(tmp, &res)
		cobra.CheckErr(err)

		// add headers
		writer = tables.NewTableWriter(command.OutOrStdout())
		headers := parseHeaders(writer, res)

		// add rows
		parseRow(writer, res, headers)
		writer.Render()
	}
}

// parseHeaders parses the headers from the result and sets them in the table
func parseHeaders(writer tables.TableOutputWriter, res map[string]interface{}) (headers []string) {
	if len(res) == 0 {
		return
	}

	for header := range res {
		headers = append(headers, header)
	}

	// add headers
	writer.SetHeaders(headers...)

	return
}

// parseRow parses the rows from the result and sets them in the table based on the headers
func parseRow(writer tables.TableOutputWriter, row map[string]interface{}, headers []string) {
	var values []interface{}

	for _, h := range headers {
		out, _ := json.MarshalIndent(row[h], "", " ")
		values = append(values, string(out))
	}

	writer.AddRow(values...)
}
