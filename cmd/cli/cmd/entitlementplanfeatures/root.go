package entitlementplanfeatures

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base feature command when called without any subcommands
var command = &cobra.Command{
	Use:   "plan-feature",
	Short: "the subcommands for working with entitlement plan features",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the planFeatures in JSON format
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetAllEntitlementPlanFeatures:
		var nodes []*openlaneclient.GetAllEntitlementPlanFeatures_EntitlementPlanFeatures_Edges_Node

		for _, i := range v.EntitlementPlanFeatures.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetEntitlementPlanFeatureByID:
		e = v.EntitlementPlanFeature
	case *openlaneclient.CreateEntitlementPlanFeature:
		e = v.CreateEntitlementPlanFeature.EntitlementPlanFeature
	case *openlaneclient.UpdateEntitlementPlanFeature:
		e = v.UpdateEntitlementPlanFeature.EntitlementPlanFeature
	case *openlaneclient.DeleteEntitlementPlanFeature:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.EntitlementPlanFeature

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.EntitlementPlanFeature
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
func tableOutput(out []openlaneclient.EntitlementPlanFeature) {
	headers := []string{"ID", "PlanName", "FeatureName"}

	// check if the planFeatures is empty and print the headers only
	if len(out) == 0 {
		writer := tables.NewTableWriter(command.OutOrStdout(), headers...)
		writer.Render()

		return
	}

	// get the metadata keys from the first planFeature and add them to the headers
	for k := range out[0].Metadata {
		headers = append(headers, k)
	}

	writer := tables.NewTableWriter(command.OutOrStdout(), headers...)

	for _, f := range out {
		items := []interface{}{f.ID, f.Plan.Name, f.Feature.Name}

		// add the metadata values to the items
		for _, v := range f.Metadata {
			items = append(items, v)
		}

		writer.AddRow(items...)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *openlaneclient.DeleteEntitlementPlanFeature) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEntitlementPlanFeature.DeletedID)

	writer.Render()
}
