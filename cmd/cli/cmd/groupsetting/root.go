//go:build cli

package groupsetting

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

// cmd represents the base group setting command when called without any subcommands
var command = &cobra.Command{
	Use:   "group-setting",
	Short: "the subcommands for working with the group settings",
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
	case *openlaneclient.GetAllGroupSettings:
		var nodes []*openlaneclient.GetAllGroupSettings_GroupSettings_Edges_Node

		for _, i := range v.GroupSettings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetGroupSettingByID:
		e = v.GroupSetting
	case *openlaneclient.UpdateGroupSetting:
		e = v.UpdateGroupSetting.GroupSetting
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.GroupSetting

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.GroupSetting
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
func tableOutput(out []openlaneclient.GroupSetting) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "GroupName", "Visibility", "SyncToGithub", "SyncToSlack")
	for _, i := range out {
		groupName := ""
		if i.Group != nil {
			groupName = i.Group.Name
		}

		writer.AddRow(i.ID,
			groupName,
			i.Visibility,
			*i.SyncToGithub,
			*i.SyncToSlack)
	}

	writer.Render()
}
