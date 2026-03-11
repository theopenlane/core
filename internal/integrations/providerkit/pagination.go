package providerkit

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

// Pagination captures common paging controls shared across provider operation configs
type Pagination struct {
	// PerPage sets the number of items to request per page
	PerPage int `json:"per_page"`
	// PageSize provides an alternate page size input
	PageSize int `json:"page_size"`
	// Page selects the page index to request
	Page int `json:"page"`
}

// EffectivePageSize returns the configured page size, falling back to defaultValue when unset
func (p Pagination) EffectivePageSize(defaultValue int) int {
	switch {
	case p.PerPage > 0:
		return p.PerPage
	case p.PageSize > 0:
		return p.PageSize
	default:
		return defaultValue
	}
}

// PayloadOptions captures optional payload inclusion controls for operation configs
type PayloadOptions struct {
	// IncludePayloads controls whether raw payloads are returned in operation results
	IncludePayloads bool `json:"include_payloads"`
}

// EnsureIncludePayloads forces include_payloads to true in a JSON config document,
// returning the original config unchanged if the update fails
func EnsureIncludePayloads(config json.RawMessage) json.RawMessage {
	out, _, err := jsonx.SetObjectKey(config, "include_payloads", true)
	if err != nil {
		return config
	}

	return out
}

// PageResult holds one page of results and the cursor for the next page
type PageResult[T any] struct {
	// Items contains the results from this page
	Items []T
	// NextToken is the cursor for the next page; an empty string signals no more pages
	NextToken string
}

// PageFetcher fetches one page of results given a page token
type PageFetcher[T any] func(ctx context.Context, pageToken string) (PageResult[T], error)

// CollectAll fetches all pages using fetch, appending results until no pages remain.
// When maxItems is greater than zero, collection stops once that many items are accumulated.
func CollectAll[T any](ctx context.Context, fetch PageFetcher[T], maxItems int) ([]T, error) {
	var (
		all       []T
		pageToken string
	)

	for {
		result, err := fetch(ctx, pageToken)
		if err != nil {
			return nil, err
		}

		all = append(all, result.Items...)

		if maxItems > 0 && len(all) >= maxItems {
			return all[:maxItems], nil
		}

		if result.NextToken == "" {
			break
		}

		pageToken = result.NextToken
	}

	return all, nil
}
