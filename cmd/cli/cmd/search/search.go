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
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)
	defer cmd.StoreSessionCookies(client)

	// filter options
	query, err := validate()
	cobra.CheckErr(err)

	results, err := client.GlobalSearch(ctx, query)
	cobra.CheckErr(err)

	return consoleOutput(results)
}

func consoleOutput(results *openlaneclient.GlobalSearch) error {
	// fragments are duplicating results with gqlgenc,
	// so we need to parse the results and create a new map
	// so it aligns with what is expected in the json output
	var realResult []map[string]interface{}

	for _, node := range results.GetSearch().GetNodes() {
		var fullResult map[string]interface{}
		err := mapstructure.Decode(node, &fullResult)
		cobra.CheckErr(err)

		for _, objectTypeResult := range fullResult {
			var parsedObjectTypeResult map[string]interface{}
			err := mapstructure.Decode(objectTypeResult, &parsedObjectTypeResult)
			cobra.CheckErr(err)

			for k, v := range parsedObjectTypeResult {
				tmp, err := json.Marshal(v)
				cobra.CheckErr(err)

				var out []interface{}
				err = json.Unmarshal(tmp, &out)
				cobra.CheckErr(err)

				if len(out) == 0 {
					continue
				}

				realResult = append(realResult, map[string]interface{}{
					k: out,
				})
			}
		}
	}

	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		// create a full result map
		full := map[string]interface{}{
			"data": map[string]interface{}{
				"search": map[string]interface{}{
					"nodes": realResult,
				},
			},
		}

		return jsonOutput(full)
	}

	tableOutput(realResult)

	return nil
}

// tableOutput prints the output in a table format
func tableOutput(results []map[string]interface{}) {
	for _, r := range results {
		writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout())

		for k, v := range r {
			// print the object type header
			fmt.Println(k)

			tmp, err := json.Marshal(v)
			cobra.CheckErr(err)

			var res []map[string]interface{}
			err = json.Unmarshal(tmp, &res)
			cobra.CheckErr(err)

			// add headers
			headers := parseHeaders(writer, res)

			// add rows
			parseRows(writer, res, headers)

			// render the table
			writer.Render()
		}
	}
}

// parseHeaders parses the headers from the result and sets them in the table
// the id column is always added as the first column
func parseHeaders(writer tables.TableOutputWriter, res []map[string]interface{}) (headers []string) {
	if len(res) == 0 {
		return
	}

	// always add the ID as the first column
	headers = append(headers, "id")

	for header := range res[0] {
		if strings.EqualFold(header, "id") {
			continue
		}

		headers = append(headers, header)
	}

	// add headers
	writer.SetHeaders(headers...)

	return
}

// parseRows parses the rows from the result and sets them in the table based on the headers
func parseRows(writer tables.TableOutputWriter, row []map[string]interface{}, headers []string) {
	for _, v := range row {
		var values []interface{}

		for _, h := range headers {
			values = append(values, fmt.Sprintf("%v", v[h]))
		}

		writer.AddRow(values...)
	}
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}
