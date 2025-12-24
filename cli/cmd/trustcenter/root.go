//go:build cli

package trustcenter

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
)

// command represents the base trustcenter command when called without any subcommands
var command = &cobra.Command{
	Use:   "trustcenter",
	Short: "the subcommands for working with trustcenters",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the trustcenters in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the trustcenters and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllTrustCenters:
		var nodes []*graphclient.GetAllTrustCenters_TrustCenters_Edges_Node

		for _, i := range v.TrustCenters.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeTrustCenters:
		var nodes []*graphclient.GeTrustCenters_TrustCenters_Edges_Node

		for _, i := range v.TrustCenters.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeTrustCenterByID:
		e = v.TrustCenter
	case *graphclient.CreateTrustCenter:
		e = v.CreateTrustCenter.TrustCenter
	case *graphclient.UpdateTrustCenter:
		e = v.UpdateTrustCenter.TrustCenter
	case *graphclient.UpdateTrustCenterSetting:
		e = v.UpdateTrustCenterSetting.TrustCenterSetting
	case *graphclient.GeTrustCenterSettingByID:
		e = v.TrustCenterSetting
	case *graphclient.DeleteTrustCenter:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	// Try to unmarshal as trust center settings first
	var setting openlane.GetTrustCenterSettingByID_TrustCenterSetting
	err = json.Unmarshal(s, &setting)
	if err == nil && setting.ID != "" {
		// This is a trust center setting
		tableSettingsOutputFromGeneric(setting)
		return nil
	}

	// Fall back to trust center list handling
	var list []graphclient.GetAllTrustCenters_TrustCenters_Edges_Node

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.GetAllTrustCenters_TrustCenters_Edges_Node
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
func tableOutput(out []graphclient.GetAllTrustCenters_TrustCenters_Edges_Node) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Slug", "CustomDomainID", "OwnerID", "Tags", "CreatedAt", "UpdatedAt")
	for _, i := range out {
		customDomainID := ""
		if i.CustomDomainID != nil {
			customDomainID = *i.CustomDomainID
		}

		ownerID := ""
		if i.OwnerID != nil {
			ownerID = *i.OwnerID
		}

		createdAt := ""
		if i.CreatedAt != nil {
			createdAt = i.CreatedAt.Format("2006-01-02 15:04:05")
		}

		updatedAt := ""
		if i.UpdatedAt != nil {
			updatedAt = i.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		writer.AddRow(i.ID, *i.Slug, customDomainID, ownerID, strings.Join(i.Tags, ","), createdAt, updatedAt)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteTrustCenter) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTrustCenter.DeletedID)

	writer.Render()
}

// tableSettingsOutputFromGeneric prints the trust center settings in a table format from generic setting type
func tableSettingsOutputFromGeneric(setting openlane.GetTrustCenterSettingByID_TrustCenterSetting) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "TrustCenterID", "Title", "Overview", "PrimaryColor", "CreatedAt", "UpdatedAt")

	title := ""
	if setting.Title != nil {
		title = *setting.Title
	}

	overview := ""
	if setting.Overview != nil {
		overview = *setting.Overview
	}

	primaryColor := ""
	if setting.PrimaryColor != nil {
		primaryColor = *setting.PrimaryColor
	}

	trustCenterID := ""
	if setting.TrustCenterID != nil {
		trustCenterID = *setting.TrustCenterID
	}

	createdAt := ""
	if setting.CreatedAt != nil {
		createdAt = setting.CreatedAt.Format("2006-01-02 15:04:05")
	}

	updatedAt := ""
	if setting.UpdatedAt != nil {
		updatedAt = setting.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	// Truncate overview if it's too long for table display
	if len(overview) > 50 {
		overview = overview[:47] + "..."
	}

	writer.AddRow(setting.ID, trustCenterID, title, overview, primaryColor, createdAt, updatedAt)

	writer.Render()
}
