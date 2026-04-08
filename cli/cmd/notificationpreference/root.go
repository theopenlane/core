//go:build cli

package notificationpreference

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base notification-preference command when called without any subcommands
var command = &cobra.Command{
	Use:   "notification-preference",
	Short: "the subcommands for working with notification preferences",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	switch v := e.(type) {
	case *graphclient.GetAllNotificationPreferences:
		var nodes []*graphclient.GetAllNotificationPreferences_NotificationPreferences_Edges_Node

		for _, i := range v.NotificationPreferences.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetNotificationPreferenceByID:
		e = v.NotificationPreference
	case *graphclient.CreateNotificationPreference:
		e = v.CreateNotificationPreference.NotificationPreference
	case *graphclient.UpdateNotificationPreference:
		e = v.UpdateNotificationPreference.NotificationPreference
	case *graphclient.DeleteNotificationPreference:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.NotificationPreference

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.NotificationPreference
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
func tableOutput(out []graphclient.NotificationPreference) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "UserID", "Channel", "Status", "Enabled", "Cadence", "Priority", "Destination", "IsDefault")
	for _, i := range out {
		writer.AddRow(i.ID, i.UserID, i.Channel, i.Status, i.Enabled, i.Cadence, derefPriority(i.Priority), derefStr(i.Destination), i.IsDefault)
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteNotificationPreference) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteNotificationPreference.DeletedID)

	writer.Render()
}

// derefStr safely dereferences a string pointer
func derefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// derefPriority safely dereferences a Priority pointer
func derefPriority(p *enums.Priority) string {
	if p == nil {
		return ""
	}

	return string(*p)
}
