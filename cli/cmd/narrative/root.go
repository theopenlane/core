//go:build cli

package narrative

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base narrative command when called without any subcommands
var command = &cobra.Command{
	Use:   "narrative",
	Short: "the subcommands for working with narratives",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the narratives in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the narratives and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllNarratives:
		var nodes []*graphclient.GetAllNarratives_Narratives_Edges_Node

		for _, i := range v.Narratives.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetNarratives:
		var nodes []*graphclient.GetNarratives_Narratives_Edges_Node

		for _, i := range v.Narratives.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetNarrativeByID:
		e = v.Narrative
	case *graphclient.CreateNarrative:
		e = v.CreateNarrative.Narrative
	case *graphclient.UpdateNarrative:
		e = v.UpdateNarrative.Narrative
	case *graphclient.DeleteNarrative:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Narrative

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Narrative
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
func tableOutput(out []graphclient.Narrative) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Details")

	for _, i := range out {
		writer.AddRow(i.ID, i.Name, *i.Description, *i.Details)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteNarrative) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteNarrative.DeletedID)

	writer.Render()
}
