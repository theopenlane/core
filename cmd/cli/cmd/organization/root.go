//go:build cli

package org

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// cmd represents the base org command when called without any subcommands
var command = &cobra.Command{
	Use:     "organization",
	Aliases: []string{"org"},
	Short:   "the subcommands for working with the organization",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllOrganizations:
		var nodes []*openlaneclient.GetAllOrganizations_Organizations_Edges_Node

		for _, i := range v.Organizations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetOrganizations:
		var nodes []*openlaneclient.GetOrganizations_Organizations_Edges_Node

		for _, i := range v.Organizations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetOrganizationByID:
		e = v.Organization
	case *openlaneclient.CreateOrganization:
		e = v.CreateOrganization.Organization
	case *openlaneclient.UpdateOrganization:
		e = v.UpdateOrganization.Organization
	case *openlaneclient.DeleteOrganization:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.Organization

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.Organization
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
func tableOutput(out []openlaneclient.Organization) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "PersonalOrg", "Children", "Members")

	for _, i := range out {
		childrenLen := 0
		if i.Children != nil {
			childrenLen = len(i.Children.Edges)
		}

		memberCount := int64(0)
		if i.Members != nil {
			memberCount = i.Members.TotalCount
		}

		writer.AddRow(i.ID,
			i.DisplayName,
			*i.Description,
			*i.PersonalOrg,
			childrenLen,
			memberCount)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteOrganization) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteOrganization.DeletedID)

	writer.Render()
}
