//go:build cli

package platform

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base platform command when called without any subcommands
var command = &cobra.Command{
	Use:   "platform",
	Short: "the subcommands for working with Platforms",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Platforms in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Platforms and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllPlatforms:
		var nodes []*graphclient.GetAllPlatforms_Platforms_Edges_Node

		for _, i := range v.Platforms.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetPlatforms:
		var nodes []*graphclient.GetPlatforms_Platforms_Edges_Node

		for _, i := range v.Platforms.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetPlatformByID:
		e = v.Platform
	case *graphclient.CreatePlatform:
		e = v.CreatePlatform.Platform
	case *graphclient.UpdatePlatform:
		e = v.UpdatePlatform.Platform
	case *graphclient.DeletePlatform:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Platform

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Platform
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
func tableOutput(out []graphclient.Platform) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Environment", "Status", "IdentityHolders")
	for _, i := range out {
		env := ""
		if i.EnvironmentName != nil {
			env = *i.EnvironmentName
		}

		identityHolderCount := 0
		if i.IdentityHolders != nil {
			identityHolderCount = len(i.IdentityHolders.Edges)
		}

		writer.AddRow(i.ID, i.DisplayID, i.Name, env, i.Status, identityHolderCount)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeletePlatform) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeletePlatform.DeletedID)

	writer.Render()
}
