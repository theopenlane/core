package microsoft

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// resourceGraphAPIVersion is the Microsoft.ResourceGraph API version used for Resource Graph queries
	resourceGraphAPIVersion = "2022-10-01"
	// resourceGraphQueryPath is the ARM path for Azure Resource Graph queries
	resourceGraphQueryPath = "/providers/Microsoft.ResourceGraph/resources?api-version=" + resourceGraphAPIVersion
	// resourceGraphPageSize is the maximum number of rows Resource Graph returns per page
	resourceGraphPageSize = 1000
	// resourceGraphResultFormat returns each row as a JSON object keyed by column name
	resourceGraphResultFormat = "objectArray"
)

// resourceGraphRequest is the Resource Graph query request body
type resourceGraphRequest struct {
	// Subscriptions scopes the query; when empty every subscription the caller can read is queried
	Subscriptions []string `json:"subscriptions,omitempty"`
	// Query is the KQL query text
	Query string `json:"query"`
	// Options controls paging and the result format
	Options resourceGraphOptions `json:"options"`
}

// resourceGraphOptions holds the Resource Graph paging and format options
type resourceGraphOptions struct {
	// ResultFormat is the row format of the response data
	ResultFormat string `json:"resultFormat"`
	// Top is the page size
	Top int `json:"$top"`
	// SkipToken continues from the previous page
	SkipToken string `json:"$skipToken,omitempty"`
}

// resourceGraphPage is one page of Resource Graph query results
type resourceGraphPage struct {
	// Data is the result rows on this page
	Data []json.RawMessage `json:"data"`
	// SkipToken continues to the next page, empty on the last page
	SkipToken string `json:"$skipToken"`
}

// QueryResourceGraph runs a KQL query against Azure Resource Graph and returns every result row
func (c *ARMClient) QueryResourceGraph(ctx context.Context, query string, subscriptionIDs []string) ([]json.RawMessage, error) {
	body := resourceGraphRequest{
		Subscriptions: subscriptionIDs,
		Query:         query,
		Options:       resourceGraphOptions{ResultFormat: resourceGraphResultFormat, Top: resourceGraphPageSize},
	}

	var rows []json.RawMessage

	for {
		page, err := fetchResourceGraphPage(ctx, c.client.Pipeline(), c.client.Endpoint()+resourceGraphQueryPath, body)
		if err != nil {
			return nil, err
		}

		rows = append(rows, page.Data...)

		if page.SkipToken == "" {
			return rows, nil
		}

		body.Options.SkipToken = page.SkipToken
	}
}

// fetchResourceGraphPage requests and decodes one page of Resource Graph query results
func fetchResourceGraphPage(ctx context.Context, pipeline runtime.Pipeline, queryURL string, body resourceGraphRequest) (resourceGraphPage, error) {
	req, err := runtime.NewRequest(ctx, http.MethodPost, queryURL)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to build resource graph request")
		return resourceGraphPage{}, ErrResourceGraphQueryFailed
	}

	if err := runtime.MarshalAsJSON(req, body); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to encode resource graph request")
		return resourceGraphPage{}, ErrResourceGraphQueryFailed
	}

	resp, err := pipeline.Do(req)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("resource graph request failed")
		return resourceGraphPage{}, ErrResourceGraphQueryFailed
	}

	defer resp.Body.Close()

	if !runtime.HasStatusCode(resp, http.StatusOK) {
		logx.FromContext(ctx).Error().Err(runtime.NewResponseError(resp)).Msg("resource graph request returned unexpected status")
		return resourceGraphPage{}, ErrResourceGraphQueryFailed
	}

	var page resourceGraphPage
	if err := runtime.UnmarshalAsJSON(resp, &page); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to decode resource graph response")
		return resourceGraphPage{}, ErrResourceGraphQueryFailed
	}

	return page, nil
}
