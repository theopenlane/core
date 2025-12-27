//go:build cli

package trustcenterwatermarkconfig

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base trustcenterwatermarkconfig command when called without any subcommands
var command = &cobra.Command{
	Use:     "trust-center-watermark-config",
	Aliases: []string{"trustcenterwatermarkconfig"},
	Short:   "the subcommands for working with trust center watermark configurations",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the watermark configs in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the watermark configs and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllTrustCenterWatermarkConfigs:
		var nodes []*graphclient.GetAllTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node

		for _, i := range v.TrustCenterWatermarkConfigs.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetTrustCenterWatermarkConfigByID:
		e = v.TrustCenterWatermarkConfig
	case *graphclient.CreateTrustCenterWatermarkConfig:
		e = v.CreateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig
	case *graphclient.UpdateTrustCenterWatermarkConfig:
		e = v.UpdateTrustCenterWatermarkConfig.TrustCenterWatermarkConfig
	case *graphclient.DeleteTrustCenterWatermarkConfig:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.GetAllTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.GetAllTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node
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
func tableOutput(out []graphclient.GetAllTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "LogoFile", "Text", "FontSize", "Opacity", "Rotation", "Color", "Font")
	for _, i := range out {
		logoFile := ""
		if i.File != nil {
			logoFile = *i.File.PresignedURL
		}
		text := ""
		if i.Text != nil {
			text = *i.Text
		}
		fontSize := 0.0
		if i.FontSize != nil {
			fontSize = *i.FontSize
		}
		opacity := 0.0
		if i.Opacity != nil {
			opacity = *i.Opacity
		}
		rotation := 0.0
		if i.Rotation != nil {
			rotation = *i.Rotation
		}
		color := ""
		if i.Color != nil {
			color = *i.Color
		}
		font := ""
		if i.Font != nil {
			font = i.Font.String()
		}
		writer.AddRow(i.ID, logoFile, text, fontSize, opacity, rotation, color, font)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteTrustCenterWatermarkConfig) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteTrustCenterWatermarkConfig.DeletedID)

	writer.Render()
}
