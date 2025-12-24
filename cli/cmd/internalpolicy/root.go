//go:build cli

package internalpolicy

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base internal policy command when called without any subcommands
var command = &cobra.Command{
	Use:   "internal-policy",
	Short: "the subcommands for working with internal policies",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the internal policies in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the internal policies and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllInternalPolicies:
		var nodes []*graphclient.GetAllInternalPolicies_InternalPolicies_Edges_Node

		for _, i := range v.InternalPolicies.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetInternalPolicies:
		var nodes []*graphclient.GetInternalPolicies_InternalPolicies_Edges_Node

		for _, i := range v.InternalPolicies.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetInternalPolicyByID:
		e = v.InternalPolicy
	case *graphclient.CreateInternalPolicy:
		e = v.CreateInternalPolicy.InternalPolicy
	case *graphclient.CreateUploadInternalPolicy:
		e = v.CreateUploadInternalPolicy.InternalPolicy
	case *graphclient.UpdateInternalPolicy:
		e = v.UpdateInternalPolicy.InternalPolicy
	case *graphclient.UpdateInternalPolicyWithFile:
		e = v.UpdateInternalPolicy.InternalPolicy
	case *graphclient.DeleteInternalPolicy:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.InternalPolicy

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.InternalPolicy
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
func tableOutput(out []graphclient.InternalPolicy) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Details", "Status", "Type", "Revision", "ReviewDue", "ReviewFrequency", "ApprovalRequired")
	for _, i := range out {
		writer.AddRow(i.ID, i.DisplayID, i.Name, *i.Details, *i.Status, *i.PolicyType, *i.Revision, *i.ReviewDue, *i.ReviewFrequency, *i.ApprovalRequired)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteInternalPolicy) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteInternalPolicy.DeletedID)

	writer.Render()
}
