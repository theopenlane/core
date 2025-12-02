//go:build cli

package usersetting

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
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
	case *openlaneclient.GetAllUserSettings:
		var nodes []*openlaneclient.GetAllUserSettings_UserSettings_Edges_Node

		for _, i := range v.UserSettings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetUserSettingByID:
		e = v.UserSetting
	case *openlaneclient.UpdateUserSetting:
		e = v.UpdateUserSetting.UserSetting
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.UserSetting

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.UserSetting
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
func tableOutput(out []openlaneclient.UserSetting) {
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
