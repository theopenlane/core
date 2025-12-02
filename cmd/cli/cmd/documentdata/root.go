//go:build cli

package documentdata

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// command represents the base document data command when called without any subcommands
var command = &cobra.Command{
	Use:   "document-data",
	Short: "the subcommands for working with document data",
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
	case *openlaneclient.GetDocumentDataByID:
		e = v.DocumentData
	case *openlaneclient.CreateDocumentData:
		e = v.CreateDocumentData.DocumentData
	case *openlaneclient.UpdateDocumentData:
		e = v.UpdateDocumentData.DocumentData
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.DocumentData

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.DocumentData
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
func tableOutput(out []openlaneclient.DocumentData) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "TemplateID", "Data", "CreatedAt", "UpdatedAt")
	for _, i := range out {
		writer.AddRow(i.ID, i.TemplateID, i.Data, *i.CreatedAt, *i.UpdatedAt)
	}

	writer.Render()
}
