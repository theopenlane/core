//go:build cli

package jobtemplate

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base jobtemplate command when called without any subcommands
var command = &cobra.Command{
	Use:   "jobtemplate",
	Short: "the subcommands for working with JobTemplates",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the JobTemplates in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the JobTemplates and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllJobTemplates:
		var nodes []*graphclient.GetAllJobTemplates_JobTemplates_Edges_Node

		for _, i := range v.JobTemplates.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeJobTemplates:
		var nodes []*graphclient.GeJobTemplates_JobTemplates_Edges_Node

		for _, i := range v.JobTemplates.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeJobTemplateByID:
		e = v.JobTemplate
	case *graphclient.CreateJobTemplate:
		e = v.CreateJobTemplate.JobTemplate
	case *graphclient.UpdateJobTemplate:
		e = v.UpdateJobTemplate.JobTemplate
	case *graphclient.DeleteJobTemplate:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlane.JobTemplate

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlane.JobTemplate
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
func tableOutput(out []openlane.JobTemplate) {
	// create a table writer
	// TODO: add additional columns to the table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID")
	for _, i := range out {
		writer.AddRow(i.ID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteJobTemplate) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteJobTemplate.DeletedID)

	writer.Render()
}
