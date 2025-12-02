//go:build cli

package organizationhistory

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// cmd represents the base organizationHistory command when called without any subcommands
var command = &cobra.Command{
	Use:     "organization-history",
	Aliases: []string{"org-history"},
	Short:   "the subcommands for working with organizationHistories",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the organizationHistories in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the organizationHistories and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllOrganizationHistories:
		var nodes []*openlaneclient.GetAllOrganizationHistories_OrganizationHistories_Edges_Node

		for _, i := range v.OrganizationHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetOrganizationHistories:
		var nodes []*openlaneclient.GetOrganizationHistories_OrganizationHistories_Edges_Node

		for _, i := range v.OrganizationHistories.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.OrganizationHistory

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.OrganizationHistory
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
func tableOutput(out []openlaneclient.OrganizationHistory) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Ref", "Operation", "UpdatedAt", "UpdatedBy")
	for _, i := range out {
		writer.AddRow(i.ID, *i.Ref, i.Operation, *i.UpdatedAt, *i.UpdatedBy)
	}

	writer.Render()
}
