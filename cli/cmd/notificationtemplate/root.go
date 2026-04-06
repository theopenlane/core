//go:build cli

package notificationtemplate

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base notification-template command when called without any subcommands
var command = &cobra.Command{
	Use:   "notification-template",
	Short: "the subcommands for working with notification templates",
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
	case *graphclient.GetAllNotificationTemplates:
		var nodes []*graphclient.GetAllNotificationTemplates_NotificationTemplates_Edges_Node

		for _, i := range v.NotificationTemplates.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetNotificationTemplateByID:
		e = v.NotificationTemplate
	case *graphclient.CreateNotificationTemplate:
		e = v.CreateNotificationTemplate.NotificationTemplate
	case *graphclient.UpdateNotificationTemplate:
		e = v.UpdateNotificationTemplate.NotificationTemplate
	case *graphclient.DeleteNotificationTemplate:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.NotificationTemplate

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.NotificationTemplate
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
func tableOutput(out []graphclient.NotificationTemplate) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Key", "Name", "Channel", "Format", "Locale", "TopicPattern", "Active", "Description")
	for _, i := range out {
		writer.AddRow(i.ID, i.Key, i.Name, i.Channel, i.Format, i.Locale, i.TopicPattern, i.Active, derefStr(i.Description))
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteNotificationTemplate) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteNotificationTemplate.DeletedID)

	writer.Render()
}

// derefStr safely dereferences a string pointer
func derefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
