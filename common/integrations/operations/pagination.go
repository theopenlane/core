package operations

import "context"

// PageResult holds one page of results and the cursor for the next page.
type PageResult[T any] struct {
	// Items contains the results from this page.
	Items []T
	// NextToken is the cursor for the next page. An empty string signals no more pages.
	NextToken string
}

// PageFetcher fetches one page of results given a page token.
// An empty pageToken requests the first page.
type PageFetcher[T any] func(ctx context.Context, pageToken string) (PageResult[T], error)

// CollectAll fetches all pages using fetch, appending results until no pages remain.
// When maxItems is greater than zero, collection stops once that many items have been gathered.
// A zero maxItems collects all available items.
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
