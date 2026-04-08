//go:build cli

package emailtemplate

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base email-template command when called without any subcommands
var command = &cobra.Command{
	Use:   "email-template",
	Short: "the subcommands for working with email templates",
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
	case *graphclient.GetAllEmailTemplates:
		var nodes []*graphclient.GetAllEmailTemplates_EmailTemplates_Edges_Node

		for _, i := range v.EmailTemplates.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetEmailTemplateByID:
		e = v.EmailTemplate
	case *graphclient.CreateEmailTemplate:
		e = v.CreateEmailTemplate.EmailTemplate
	case *graphclient.UpdateEmailTemplate:
		e = v.UpdateEmailTemplate.EmailTemplate
	case *graphclient.DeleteEmailTemplate:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.EmailTemplate

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.EmailTemplate
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
func tableOutput(out []graphclient.EmailTemplate) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Key", "Name", "Format", "Locale", "TemplateContext", "Active", "Version", "Description")
	for _, i := range out {
		writer.AddRow(i.ID, i.Key, i.Name, i.Format, i.Locale, i.TemplateContext, i.Active, i.Version, derefStr(i.Description))
	}

	writer.Render()
}

// deletedTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteEmailTemplate) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteEmailTemplate.DeletedID)

	writer.Render()
}

// derefStr safely dereferences a string pointer
func derefStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
