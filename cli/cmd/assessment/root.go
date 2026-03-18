//go:build cli

package assessment

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base assessment command when called without any subcommands
var command = &cobra.Command{
	Use:   "assessment",
	Short: "the subcommands for working with Assessments",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Assessments in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Assessments and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllAssessments:
		var nodes []*graphclient.GetAllAssessments_Assessments_Edges_Node

		for _, i := range v.Assessments.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetAssessments:
		var nodes []*graphclient.GetAssessments_Assessments_Edges_Node

		for _, i := range v.Assessments.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetAssessmentByID:
		e = v.Assessment
	case *graphclient.CreateAssessment:
		e = v.CreateAssessment.Assessment
	case *graphclient.UpdateAssessment:
		e = v.UpdateAssessment.Assessment
	case *graphclient.DeleteAssessment:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Assessment

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Assessment
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
func tableOutput(out []graphclient.Assessment) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Type", "Campaigns", "CreatedAt")
	for _, i := range out {
		createdAt := ""
		if i.CreatedAt != nil {
			createdAt = i.CreatedAt.Format("2006-01-02 15:04")
		}

		campaignCount := 0
		if i.Campaigns != nil {
			campaignCount = len(i.Campaigns.Edges)
		}

		writer.AddRow(i.ID, i.Name, i.AssessmentType, campaignCount, createdAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteAssessment) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteAssessment.DeletedID)

	writer.Render()
}
