//go:build cli

package usersetting

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// cmd represents the base user setting command when called without any subcommands
var command = &cobra.Command{
	Use:   "user-setting",
	Short: "the subcommands for working with the user settings",
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
	case *graphclient.GetAllUserSettings:
		var nodes []*graphclient.GetAllUserSettings_UserSettings_Edges_Node

		for _, i := range v.UserSettings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetUserSettingByID:
		e = v.UserSetting
	case *graphclient.UpdateUserSetting:
		e = v.UpdateUserSetting.UserSetting
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.UserSetting

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.UserSetting
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
func tableOutput(out []graphclient.UserSetting) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DefaultOrgName", "DefaultOrgID", "2FA", "Status", "EmailConfirmed", "Tags")

	for _, i := range out {
		defaultOrgName := ""
		defaultOrgID := ""

		if i.DefaultOrg != nil {
			defaultOrgName = i.DefaultOrg.DisplayName
			defaultOrgID = i.DefaultOrg.ID
		}

		writer.AddRow(i.ID, defaultOrgName, defaultOrgID, *i.IsTfaEnabled, i.Status, i.EmailConfirmed, strings.Join(i.Tags, ", "))
	}

	writer.Render()
}
