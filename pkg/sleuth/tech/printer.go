package tech

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/rs/zerolog/log"
)

// PrintAppInfoTable prints the AppInfo in a table format with colorized output and fixed-width columns
// this is a helper function for getting tidy results into the console, but the actual output we'd have in production would be json
func PrintAppInfoTable(appInfo map[string]AppInfo) {
	if len(appInfo) == 0 {
		log.Warn().Msg("No AppInfo data to display")
		return
	}

	maxColumnWidth := 50

	opts := []tablewriter.Option{
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{})),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
			Row: tw.CellConfig{
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapTruncate}, // truncate long content
				ColMaxWidths: tw.CellWidth{Global: maxColumnWidth},
			},
			Footer: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
		}),
	}

	table := tablewriter.NewTable(os.Stdout, opts...)
	table.Header([]string{"Name", "Description", "Website", "CPE", "Icon", "Categories"})

	for name, info := range appInfo {
		row := []string{
			name, info.Description, info.Website, info.CPE, info.Icon, fmt.Sprintf("%v", info.Categories),
		}

		_ = table.Append(row)
	}

	_ = table.Render()
}
