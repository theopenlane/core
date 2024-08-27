package entitlement

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base entitlement command when called without any subcommands
var command = &cobra.Command{
	Use:   "entitlement",
	Short: "the subcommands for working with entitlements",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetEntitlements:
		var nodes []*openlaneclient.GetEntitlements_Entitlements_Edges_Node

		for _, i := range v.Entitlements.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetEntitlementByID:
		e = v.Entitlement
	case *openlaneclient.CreateEntitlement:
		e = v.CreateEntitlement.Entitlement
	case *openlaneclient.UpdateEntitlement:
		e = v.UpdateEntitlement.Entitlement
	case *openlaneclient.DeleteEntitlement:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var entitlementList []openlaneclient.Entitlement

	err = json.Unmarshal(s, &entitlementList)
	if err != nil {
		var entitlement openlaneclient.Entitlement
		err = json.Unmarshal(s, &entitlement)
		cobra.CheckErr(err)

		entitlementList = append(entitlementList, entitlement)
	}

	tableOutput(entitlementList)

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the entitlements in a table format
func tableOutput(out []openlaneclient.Entitlement) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "PlanID", "PlanName", "OrganizationID", "OrganizationName", "ExpiresAt", "Expires", "Cancelled")

	for _, i := range out {
		writer.AddRow(i.ID, i.Plan.ID, i.Plan.Name, i.Organization.ID, i.Organization.Name, i.ExpiresAt, i.Expires, i.Cancelled)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteEntitlement) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEntitlement.DeletedID)

	writer.Render()
}
