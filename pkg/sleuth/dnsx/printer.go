package dnsx

import (
	"fmt"
	"os"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

// PrintDNSRecordsReportTable prints the DNS records report in a table format with colorized output and fixed-width columns
func PrintDNSRecordsReportTable(report DNSRecordsReport) {
	width, _, err := term.GetSize(int(os.Stdout.Fd())) // nolint:gosec
	if err != nil {
		width = 80 // default width if terminal size cannot be determined
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Name", "Value", "TTL", "CDN"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	maxColumnWidth := width / 4 // nolint:mnd

	addRecordsToTable := func(records []*DNSRecord) {
		for _, record := range records {
			row := []string{
				runewidth.Truncate(record.Type, maxColumnWidth, "..."),
				runewidth.Truncate(record.Name, maxColumnWidth, "..."),
				runewidth.Truncate(record.Value, maxColumnWidth, "..."),
				runewidth.Truncate(fmt.Sprintf("%d", record.TTL), maxColumnWidth, "..."),
				runewidth.Truncate(record.CDN, maxColumnWidth, "..."),
			}
			table.Append(row)
		}
	}

	if report.DNSRecords != nil {
		addRecordsToTable(report.DNSRecords.A)
		addRecordsToTable(report.DNSRecords.AAAA)
		addRecordsToTable(report.DNSRecords.MX)
		addRecordsToTable(report.DNSRecords.Txt)
		addRecordsToTable(report.DNSRecords.NS)
		addRecordsToTable(report.DNSRecords.CNAME)
	}

	if len(report.Errors) > 0 {
		pterm.Error.Println("Errors:")

		for _, err := range report.Errors {
			pterm.Error.Println(err)
		}
	}

	table.Render()
}
