//go:build cli

package subprocessor

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base subprocessor command when called without any subcommands
var command = &cobra.Command{
	Use:   "subprocessor",
	Short: "the subcommands for working with subprocessors",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the subprocessors in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the subprocessors and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllSubprocessors:
		var nodes []*openlaneclient.GetAllSubprocessors_Subprocessors_Edges_Node

		for _, i := range v.Subprocessors.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetSubprocessors:
		var nodes []*openlaneclient.GetSubprocessors_Subprocessors_Edges_Node

		for _, i := range v.Subprocessors.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetSubprocessorByID:
		e = v.Subprocessor
	case *openlaneclient.CreateSubprocessor:
		e = v.CreateSubprocessor.Subprocessor
	case *openlaneclient.UpdateSubprocessor:
		e = v.UpdateSubprocessor.Subprocessor
	case *openlaneclient.DeleteSubprocessor:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Subprocessor

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Subprocessor
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
func tableOutput(out []openlaneclient.Subprocessor) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Logo Remote URL", "Owner ID")
	for _, i := range out {
		description := ""
		if i.Description != nil {
			description = *i.Description
		}
		logoURL := ""
		if i.LogoRemoteURL != nil {
			logoURL = *i.LogoRemoteURL
		}
		ownerID := ""
		if i.OwnerID != nil {
			ownerID = *i.OwnerID
		}
		writer.AddRow(i.ID, i.Name, description, logoURL, ownerID)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteSubprocessor) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteSubprocessor.DeletedID)

	writer.Render()
}
