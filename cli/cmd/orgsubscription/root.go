//go:build cli

package orgsubscription

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base orgSubscription command when called without any subcommands
var command = &cobra.Command{
	Use:     "org-subscription",
	Aliases: []string{"organization-subscription"},
	Short:   "the subcommands for working with orgSubscriptions",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the orgSubscriptions in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the orgSubscriptions and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllOrgSubscriptions:
		var nodes []*graphclient.GetAllOrgSubscriptions_OrgSubscriptions_Edges_Node

		for _, i := range v.OrgSubscriptions.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetOrgSubscriptions:
		var nodes []*graphclient.GetOrgSubscriptions_OrgSubscriptions_Edges_Node

		for _, i := range v.OrgSubscriptions.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetOrgSubscriptionByID:
		e = v.OrgSubscription
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.OrgSubscription

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.OrgSubscription
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
func tableOutput(out []graphclient.OrgSubscription) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Active", "StripeSubscriptionStatus", "ExpiresAt")
	for _, i := range out {
		exp := ""
		if i.ExpiresAt != nil {
			exp = i.ExpiresAt.String()
		} else if i.TrialExpiresAt != nil {
			exp = i.TrialExpiresAt.String()
		}

		writer.AddRow(i.ID,
			i.Active,
			*i.StripeSubscriptionStatus,
			exp,
		)
	}

	writer.Render()
}
