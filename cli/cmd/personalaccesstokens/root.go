//go:build cli

package tokens

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
)

// cmd represents the base cmd command when called without any subcommands
var command = &cobra.Command{
	Use:   "pat",
	Short: "the subcommands for working with personal access tokens",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the pat in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the pat and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllPersonalAccessTokens:
		var nodes []*graphclient.GetAllPersonalAccessTokens_PersonalAccessTokens_Edges_Node

		for _, i := range v.PersonalAccessTokens.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GePersonalAccessTokenByID:
		e = v.PersonalAccessToken
	case *graphclient.CreatePersonalAccessToken:
		e = v.CreatePersonalAccessToken.PersonalAccessToken
	case *graphclient.UpdatePersonalAccessToken:
		e = v.UpdatePersonalAccessToken.PersonalAccessToken
	case *graphclient.DeletePersonalAccessToken:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlane.PersonalAccessToken

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlane.PersonalAccessToken
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
func tableOutput(out []openlane.PersonalAccessToken) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Token", "Authorized Organizations", "LastUsedAt", "ExpiresAt")

	for _, i := range out {
		lastUsed := "never"
		if i.LastUsedAt != nil {
			lastUsed = i.LastUsedAt.String()
		}

		expiresAt := "never"
		if i.ExpiresAt != nil {
			expiresAt = i.ExpiresAt.String()
		}

		orgs := []string{}
		for _, o := range i.Organizations.Edges {
			orgs = append(orgs, o.Node.Name)
		}

		authorizedOrgs := strings.Join(orgs, ", ")

		writer.AddRow(i.ID, i.Name, i.Token, authorizedOrgs, lastUsed, expiresAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeletePersonalAccessToken) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeletePersonalAccessToken.DeletedID)

	writer.Render()
}
