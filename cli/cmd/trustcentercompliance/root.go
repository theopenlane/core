//go:build cli

package trustcentercompliance

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base trustcentercompliance command when called without any subcommands
var command = &cobra.Command{
	Use:   "trustcentercompliance",
	Short: "the subcommands for working with trust center compliances",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the trust center compliances in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check if the output is a slice of trust center compliances
	if trustCenterCompliances, ok := e.(*graphclient.GetAllTrustCenterCompliances); ok {
		var nodes []*graphclient.GetAllTrustCenterCompliances_TrustCenterCompliances_Edges_Node

		for _, edge := range trustCenterCompliances.TrustCenterCompliances.Edges {
			nodes = append(nodes, edge.Node)
		}

		e = nodes
	}

	// check if the output is a single trust center compliance
	if trustCenterCompliance, ok := e.(*graphclient.GetTrustCenterComplianceByID); ok {
		e = trustCenterCompliance.TrustCenterCompliance
	}

	// check if the output is a create trust center compliance response
	if createResp, ok := e.(*graphclient.CreateTrustCenterCompliance); ok {
		e = createResp.CreateTrustCenterCompliance.TrustCenterCompliance
	}

	// check if the output is a delete trust center compliance response
	if deleteResp, ok := e.(*graphclient.DeleteTrustCenterCompliance); ok {
		deletedTableOutput(deleteResp)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.GetAllTrustCenterCompliances_TrustCenterCompliances_Edges_Node

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.GetAllTrustCenterCompliances_TrustCenterCompliances_Edges_Node
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
func tableOutput(out []graphclient.GetAllTrustCenterCompliances_TrustCenterCompliances_Edges_Node) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "STANDARD", "TAGS", "CREATED", "UPDATED")
	for _, i := range out {
		writer.AddRow(i.ID, i.Standard.Name, strings.Join(i.Tags, ", "), *i.CreatedAt, *i.UpdatedAt)
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteTrustCenterCompliance) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTrustCenterCompliance.DeletedID)

	writer.Render()
}
