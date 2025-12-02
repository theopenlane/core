//go:build cli

package apitokens

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// cmd represents the base cmd command when called without any subcommands
var command = &cobra.Command{
	Use:   "token",
	Short: "the subcommands for working with api tokens",
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
	case *openlaneclient.GetAllAPITokens:
		var nodes []*openlaneclient.GetAllAPITokens_APITokens_Edges_Node

		for _, i := range v.APITokens.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetAPITokenByID:
		e = v.APIToken
	case *openlaneclient.CreateAPIToken:
		e = v.CreateAPIToken.APIToken
	case *openlaneclient.UpdateAPIToken:
		e = v.UpdateAPIToken.APIToken
	case *openlaneclient.DeleteAPIToken:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.APIToken

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.APIToken
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
func tableOutput(out []openlaneclient.APIToken) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Token", "Scopes", "LastUsedAt", "ExpiresAt")

	for _, i := range out {
		lastUsed := "never"
		if i.LastUsedAt != nil {
			lastUsed = i.LastUsedAt.String()
		}

		expiresAt := "never"
		if i.ExpiresAt != nil {
			expiresAt = i.ExpiresAt.String()
		}

		writer.AddRow(i.ID, i.Name, i.Token, strings.Join(i.Scopes, ", "), lastUsed, expiresAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteAPIToken) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteAPIToken.DeletedID)

	writer.Render()
}
