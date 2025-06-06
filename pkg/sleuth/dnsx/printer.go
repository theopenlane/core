package dnsx

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/pterm/pterm"
)

// PrintDNSRecordsReportTable prints the DNS records report in a table format with colorized output and fixed-width columns
func PrintDNSRecordsReportTable(report DNSRecordsReport) {
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
	table.Header([]string{"Type", "Name", "Value", "TTL", "CDN"})

	addRecordsToTable := func(records []*DNSRecord) {
		for _, record := range records {
			row := []string{record.Type, record.Name, record.Value, fmt.Sprintf("%d", record.TTL), record.CDN}

			_ = table.Append(row)
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

	_ = table.Render()
}
