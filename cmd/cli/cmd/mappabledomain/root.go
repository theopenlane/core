//go:build cli

package mappabledomain

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base mappabledomain command when called without any subcommands
var command = &cobra.Command{
	Use:     "mappable-domain",
	Aliases: []string{"mappabledomain", "md"},
	Short:   "the subcommands for working with mappable domains",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the mappable domains in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// otherwise print the mappable domains in a table
	var nodes []map[string]interface{}

	switch v := e.(type) {
	case *openlaneclient.GetAllMappableDomains:
		for _, edge := range v.MappableDomains.Edges {
			nodeMap := map[string]interface{}{
				"ID":        edge.Node.ID,
				"Name":      edge.Node.Name,
				"CreatedAt": edge.Node.CreatedAt,
				"CreatedBy": edge.Node.CreatedBy,
			}
			nodes = append(nodes, nodeMap)
		}
		e = nodes
	case *openlaneclient.GetMappableDomains:
		for _, edge := range v.MappableDomains.Edges {
			nodeMap := map[string]interface{}{
				"ID":        edge.Node.ID,
				"Name":      edge.Node.Name,
				"CreatedAt": edge.Node.CreatedAt,
				"CreatedBy": edge.Node.CreatedBy,
			}
			nodes = append(nodes, nodeMap)
		}
		e = nodes
	case *openlaneclient.GetMappableDomainByID:
		e = v.MappableDomain
	case *openlaneclient.CreateMappableDomain:
		e = v.CreateMappableDomain.MappableDomain
	case *openlaneclient.UpdateMappableDomain:
		e = v.UpdateMappableDomain.MappableDomain
	case *openlaneclient.DeleteMappableDomain:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.MappableDomain

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.MappableDomain
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
func tableOutput(out interface{}) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "ZoneID", "Created At", "Created By")

	switch v := out.(type) {
	case []map[string]interface{}:
		for _, i := range v {
			writer.AddRow(i["ID"], i["Name"], i["ZoneID"], i["CreatedAt"], i["CreatedBy"])
		}
	case openlaneclient.MappableDomain:
		writer.AddRow(v.ID, v.Name, v.ZoneID, v.CreatedAt, v.CreatedBy)
	case []openlaneclient.MappableDomain:
		for _, i := range v {
			cb := ""
			if i.CreatedBy != nil {
				cb = *i.CreatedBy
			}
			writer.AddRow(i.ID, i.Name, i.ZoneID, i.CreatedAt, cb)
		}
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteMappableDomain) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteMappableDomain.DeletedID)

	writer.Render()
}
