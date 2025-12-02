//go:build cli

package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var command = &cobra.Command{
	Use:   "search",
	Short: "search for organizations, groups, users, subscribers, etc in the system",
	Run: func(cmd *cobra.Command, args []string) {
		err := search(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().StringP("query", "q", "", "query string to search for")
}

// validate validates the required fields for the command
func validate() (string, error) {
	query := cmd.Config.String("query")
	if query == "" {
		return "", cmd.NewRequiredFieldMissingError("query")
	}

	return query, nil
}

// search searches for organizations, groups, users, subscribers, etc in the system
func search(ctx context.Context) error { // setup http client
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	// filter options
	query, err := validate()
	cobra.CheckErr(err)

	results, err := client.GlobalSearch(ctx, query)
	cobra.CheckErr(err)

	return consoleOutput(results)
}

func consoleOutput(results *openlaneclient.GlobalSearch) error {
	var fullResult map[string]any

	err := mapstructure.Decode(results, &fullResult)
	cobra.CheckErr(err)

	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(fullResult)
	}

	tableOutput(fullResult)

	return nil
}

// tableOutput prints the output in a table format
func tableOutput(results map[string]any) {

	for _, r := range results {

		tmp, err := json.Marshal(r)
		cobra.CheckErr(err)

		var res map[string]any
		err = json.Unmarshal(tmp, &res)
		cobra.CheckErr(err)

		for k, v := range res {
			// setup the table writer per object type
			writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout())

			// skip the totalCount field
			if strings.EqualFold(k, "totalCount") {
				continue
			}

			// print the object type header
			fmt.Println(strings.ToUpper(k))

			edge, ok := v.(map[string]any)
			if !ok {
				continue
			}

			nodes, ok := edge["edges"].([]any)
			if !ok {
				continue
			}

			headers := make([]string, len(nodes))

			for i, node := range nodes {
				n, ok := node.(map[string]any)
				if !ok {
					continue
				}

				tmp, err := json.Marshal(n["node"])
				cobra.CheckErr(err)

				var res map[string]any
				err = json.Unmarshal(tmp, &res)
				cobra.CheckErr(err)

				// add headers the first time

				if i == 0 {
					headers = parseHeaders(writer, res)
				}

				// add rows
				parseRows(writer, res, headers)

			}

			// render the table
			writer.Render()
		}

	}
}

// parseHeaders parses the headers from the result and sets them in the table
// the id column is always added as the first column
func parseHeaders(writer tables.TableOutputWriter, res map[string]any) []string {
	if len(res) == 0 {
		return nil
	}

	headers := make([]string, len(res))

	// always add the ID as the first column
	headers[0] = "ID"

	// add other fields with ordering correctly
	i := 1
	for k, _ := range res {
		if strings.EqualFold(k, "id") {
			continue
		}

		headers[i] = k

		i++
	}

	// add empty row
	writer.AddRow()

	// add headers
	writer.SetHeaders(headers...)

	return headers
}

func parseRows(writer tables.TableOutputWriter, res map[string]any, headers []string) {
	if len(res) == 0 {
		return
	}

	values := make([]any, len(res))

	// always add the ID as the first column
	values[0] = res["id"]

	// add other fields with ordering correctly
	for i, h := range headers {
		if strings.EqualFold(h, "id") {
			continue
		}

		values[i] = res[h]

		i++
	}

	writer.AddRow(values...)

	return
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}
