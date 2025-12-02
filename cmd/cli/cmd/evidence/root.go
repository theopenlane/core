//go:build cli

package evidence

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base evidence command when called without any subcommands
var command = &cobra.Command{
	Use:   "evidence",
	Short: "the subcommands for working with evidences",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the evidences in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the evidences and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllEvidences:
		var nodes []*openlaneclient.GetAllEvidences_Evidences_Edges_Node

		for _, i := range v.Evidences.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetEvidences:
		var nodes []*openlaneclient.GetEvidences_Evidences_Edges_Node

		for _, i := range v.Evidences.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetEvidenceByID:
		e = v.Evidence
	case *openlaneclient.CreateEvidence:
		e = v.CreateEvidence.Evidence
	case *openlaneclient.UpdateEvidence:
		e = v.UpdateEvidence.Evidence
	case *openlaneclient.DeleteEvidence:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Evidence

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Evidence
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
func tableOutput(out []openlaneclient.Evidence) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Description", "CollectionProcedure", "Source", "IsAutomated", "URL", "NumOfFiles", "Controls", "CreatedAt", "UpdatedAt")
	for _, i := range out {
		files := 0
		if i.Files != nil {
			files = len(i.Files.Edges)
		}

		controls := []string{}
		if i.Controls != nil {
			for _, c := range i.Controls.Edges {
				controls = append(controls, c.Node.RefCode)
			}
		}
		writer.AddRow(i.ID, i.DisplayID, i.Name, *i.Description, *i.CollectionProcedure, *i.Source, *i.IsAutomated, *i.URL, files, strings.Join(controls, ", "), i.CreatedAt, i.UpdatedAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteEvidence) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEvidence.DeletedID)

	writer.Render()
}
