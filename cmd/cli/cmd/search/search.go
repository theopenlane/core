//go:build cli

package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildSearchQuery() (string, error) {
	query := cmd.Config.String("query")
	if strings.TrimSpace(query) == "" {
		return "", cmd.NewRequiredFieldMissingError("query")
	}

	return query, nil
}

func executeSearch(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.GlobalSearch, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	query, err := buildSearchQuery()
	if err != nil {
		return nil, err
	}

	return client.GlobalSearch(ctx, query)
}

func renderSearchResults(results *openlaneclient.GlobalSearch) error {
	var fullResult map[string]any

	if err := mapstructure.Decode(results, &fullResult); err != nil {
		return err
	}

	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return speccli.PrintJSON(fullResult)
	}

	return tableOutput(fullResult)
}

func tableOutput(results map[string]any) error {
	for _, r := range results {
		tmp, err := json.Marshal(r)
		if err != nil {
			return err
		}

		var res map[string]any
		if err := json.Unmarshal(tmp, &res); err != nil {
			return err
		}

		for k, v := range res {
			writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout())
			if strings.EqualFold(k, "totalCount") {
				continue
			}

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
				if err != nil {
					return err
				}

				var res map[string]any
				if err := json.Unmarshal(tmp, &res); err != nil {
					return err
				}

				if i == 0 {
					headers = parseHeaders(writer, res)
				}

				parseRows(writer, res, headers)
			}

			writer.Render()
		}
	}

	return nil
}

func parseHeaders(writer tables.TableOutputWriter, res map[string]any) []string {
	if len(res) == 0 {
		return nil
	}

	headers := make([]string, len(res))
	headers[0] = "ID"

	i := 1
	for k := range res {
		if strings.EqualFold(k, "id") {
			continue
		}

		headers[i] = k
		i++
	}

	writer.AddRow()
	writer.SetHeaders(headers...)

	return headers
}

func parseRows(writer tables.TableOutputWriter, res map[string]any, headers []string) {
	if len(res) == 0 {
		return
	}

	values := make([]any, len(headers))
	values[0] = res["id"]

	for i, h := range headers {
		if strings.EqualFold(h, "id") {
			continue
		}

		values[i] = res[h]
	}

	writer.AddRow(values...)
}
