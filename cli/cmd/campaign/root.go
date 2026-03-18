//go:build cli

package campaign

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base campaign command when called without any subcommands
var command = &cobra.Command{
	Use:   "campaign",
	Short: "the subcommands for working with Campaigns",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Campaigns in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Campaigns and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllCampaigns:
		var nodes []*graphclient.GetAllCampaigns_Campaigns_Edges_Node

		for _, i := range v.Campaigns.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetCampaigns:
		var nodes []*graphclient.GetCampaigns_Campaigns_Edges_Node

		for _, i := range v.Campaigns.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetCampaignByID:
		e = v.Campaign
	case *graphclient.CreateCampaign:
		e = v.CreateCampaign.Campaign
	case *graphclient.UpdateCampaign:
		e = v.UpdateCampaign.Campaign
	case *graphclient.DeleteCampaign:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Campaign

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Campaign
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
func tableOutput(out []graphclient.Campaign) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayID", "Name", "Status", "Type", "Active", "Targets")
	for _, i := range out {
		targetCount := 0
		if i.CampaignTargets != nil {
			targetCount = len(i.CampaignTargets.Edges)
		}

		writer.AddRow(i.ID, i.DisplayID, i.Name, i.Status, i.CampaignType, i.IsActive, targetCount)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteCampaign) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteCampaign.DeletedID)

	writer.Render()
}

// consoleOutputWithTargets prints a single campaign with target details
func consoleOutputWithTargets(e *graphclient.GetCampaignByID) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	tableOutputCampaignWithTargets(e.Campaign.ID, e.Campaign.DisplayID, e.Campaign.Name, e.Campaign.CampaignTargets.Edges)

	return nil
}

// consoleOutputWithTargetsAll prints all campaigns with target details
func consoleOutputWithTargetsAll(e *graphclient.GetAllCampaigns) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	writer := tables.NewTableWriter(command.OutOrStdout(), "CampaignID", "CampaignName", "TargetEmail", "TargetName", "Status")
	for _, edge := range e.Campaigns.Edges {
		c := edge.Node
		targets := c.CampaignTargets.Edges
		if len(targets) == 0 {
			writer.AddRow(c.ID, c.Name, "-", "-", "-")
			continue
		}

		for idx, targetEdge := range targets {
			t := targetEdge.Node
			fullName := "-"
			if t.FullName != nil {
				fullName = *t.FullName
			}

			if idx == 0 {
				writer.AddRow(c.ID, c.Name, t.Email, fullName, t.Status)
			} else {
				writer.AddRow("", "", t.Email, fullName, t.Status)
			}
		}
	}

	writer.Render()

	return nil
}

// tableOutputCampaignWithTargets prints a single campaign with its targets
func tableOutputCampaignWithTargets(campaignID, displayID, name string, targets []*graphclient.GetCampaignByID_Campaign_CampaignTargets_Edges) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "CampaignID", "DisplayID", "Name", "TargetEmail", "TargetName", "Status")
	if len(targets) == 0 {
		writer.AddRow(campaignID, displayID, name, "-", "-", "-")
		writer.Render()

		return
	}

	for idx, targetEdge := range targets {
		t := targetEdge.Node
		fullName := "-"
		if t.FullName != nil {
			fullName = *t.FullName
		}

		if idx == 0 {
			writer.AddRow(campaignID, displayID, name, t.Email, fullName, t.Status)
		} else {
			writer.AddRow("", "", "", t.Email, fullName, t.Status)
		}
	}

	writer.Render()
}
