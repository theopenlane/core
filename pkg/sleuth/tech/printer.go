package tech

import (
	"fmt"
	"os"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

// PrintAppInfoTable prints the AppInfo in a table format with colorized output and fixed-width columns
// this is a helper function for getting tidy results into the console, but the actual output we'd have in production would be json
func PrintAppInfoTable(appInfo map[string]AppInfo) {
	if len(appInfo) == 0 {
		log.Warn().Msg("No AppInfo data to display")
		return
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd())) // nolint:gosec
	if err != nil {
		width = 80 // default width if terminal size cannot be determined
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "Website", "CPE", "Icon", "Categories"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	maxColumnWidth := width / 6 // nolint:mnd

	for name, info := range appInfo {
		row := []string{
			runewidth.Truncate(name, maxColumnWidth, "..."),
			runewidth.Truncate(info.Description, maxColumnWidth, "..."),
			runewidth.Truncate(info.Website, maxColumnWidth, "..."),
			runewidth.Truncate(info.CPE, maxColumnWidth, "..."),
			runewidth.Truncate(info.Icon, maxColumnWidth, "..."),
			runewidth.Truncate(fmt.Sprintf("%v", info.Categories), maxColumnWidth, "..."),
		}
		table.Append(row)
	}

	table.Render()
}
