package evidence

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Description", "CollectionProcedure", "Source", "IsAutomated", "URL", "NumOfFiles", "Programs", "CreatedAt", "UpdatedAt")
	for _, i := range out {
		prgs := i.Programs

		progNames := make([]string, 0, len(prgs))

		// add programs with either no end date or end date before the current date
		for _, p := range prgs {
			if p.EndDate == nil {
				progNames = append(progNames, p.Name)
			} else if p.EndDate.Before(time.Now()) {
				progNames = append(progNames, p.Name)
			}
		}

		writer.AddRow(i.ID, i.DisplayID, i.Name, *i.Description, *i.CollectionProcedure, *i.Source, *i.IsAutomated, *i.URL, len(i.Files), strings.Join(progNames, ","), i.CreatedAt, i.UpdatedAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteEvidence) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEvidence.DeletedID)

	writer.Render()
}
