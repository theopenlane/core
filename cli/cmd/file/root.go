//go:build cli

package file

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
)

// command represents the base file command when called without any subcommands
var command = &cobra.Command{
	Use:   "file",
	Short: "the subcommands for working with files",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the files in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the files and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllFiles:
		var nodes []*graphclient.GetAllFiles_Files_Edges_Node

		for _, i := range v.Files.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeFiles:
		var nodes []*graphclient.GeFiles_Files_Edges_Node

		for _, i := range v.Files.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GeFileByID:
		e = v.File
	case *graphclient.DeleteFile:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlane.File

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlane.File
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
func tableOutput(out []openlane.File) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "URI", "ProvidedFileName", "CategoryType", "DetectedMimeType", "FileExtension", "StoragePath")
	for _, i := range out {
		writer.AddRow(i.ID, *i.URI, i.ProvidedFileName, *i.Md5Hash, *i.DetectedMimeType, i.ProvidedFileExtension, *i.StoragePath)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteFile) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteFile.DeletedID)

	writer.Render()
}
