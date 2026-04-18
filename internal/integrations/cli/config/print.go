package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/samber/lo"
)

// tabWriterPadding is the padding between table columns
const tabWriterPadding = 3

// JSONOutput prints the value as indented JSON without HTML escaping
func JSONOutput(out any) error {
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(out); err != nil {
		return err
	}

	_, err := fmt.Print(buf.String())

	return err
}

// TableOutput prints a simple tab-delimited table to stdout
func TableOutput(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabWriterPadding, ' ', 0)

	if len(headers) > 0 {
		fmt.Fprintln(w, strings.Join(headers, "\t"))

		separators := lo.Map(headers, func(h string, _ int) string {
			return strings.Repeat("-", len(h))
		})

		fmt.Fprintln(w, strings.Join(separators, "\t"))
	}

	for _, row := range rows {
		if len(headers) > 0 && len(row) < len(headers) {
			padded := make([]string, len(headers))
			copy(padded, row)
			row = padded
		}

		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}
