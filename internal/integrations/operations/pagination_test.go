package operations_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/operations"
)

func TestCollectAll_SinglePage(t *testing.T) {
	fetch := func(_ context.Context, _ string) (operations.PageResult[int], error) {
		return operations.PageResult[int]{Items: []int{1, 2, 3}}, nil
	}

	got, err := operations.CollectAll(context.Background(), fetch, 0)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, got)
}

func TestCollectAll_MultiplePages(t *testing.T) {
	pages := map[string]operations.PageResult[int]{
		"":      {Items: []int{1, 2}, NextToken: "page2"},
		"page2": {Items: []int{3, 4}, NextToken: "page3"},
		"page3": {Items: []int{5, 6}},
	}

	fetch := func(_ context.Context, token string) (operations.PageResult[int], error) {
		return pages[token], nil
	}

	got, err := operations.CollectAll(context.Background(), fetch, 0)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, got)
}

func TestCollectAll_MaxItemsCap(t *testing.T) {
	pages := map[string]operations.PageResult[int]{
		"":      {Items: []int{1, 2, 3}, NextToken: "page2"},
		"page2": {Items: []int{4, 5, 6}},
	}

	fetch := func(_ context.Context, token string) (operations.PageResult[int], error) {
		return pages[token], nil
	}

	got, err := operations.CollectAll(context.Background(), fetch, 4)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4}, got)
}

func TestCollectAll_MaxItemsExactPage(t *testing.T) {
	callCount := 0
	pages := map[string]operations.PageResult[int]{
		"":      {Items: []int{1, 2, 3}, NextToken: "page2"},
		"page2": {Items: []int{4, 5, 6}},
	}

	fetch := func(_ context.Context, token string) (operations.PageResult[int], error) {
		callCount++
		return pages[token], nil
	}

	got, err := operations.CollectAll(context.Background(), fetch, 3)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, got)
	assert.Equal(t, 1, callCount, "should stop after first page when maxItems reached")
}

func TestCollectAll_Empty(t *testing.T) {
	fetch := func(_ context.Context, _ string) (operations.PageResult[int], error) {
		return operations.PageResult[int]{}, nil
	}

	got, err := operations.CollectAll(context.Background(), fetch, 0)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCollectAll_FetchError(t *testing.T) {
	fetchErr := errors.New("fetch failed")
	fetch := func(_ context.Context, _ string) (operations.PageResult[int], error) {
		return operations.PageResult[int]{}, fetchErr
	}

	_, err := operations.CollectAll(context.Background(), fetch, 0)
	assert.ErrorIs(t, err, fetchErr)
}

func TestCollectAll_ErrorOnSecondPage(t *testing.T) {
	fetchErr := errors.New("second page failed")
	callCount := 0

	fetch := func(_ context.Context, token string) (operations.PageResult[int], error) {
		callCount++
		if callCount == 1 {
			return operations.PageResult[int]{Items: []int{1, 2}, NextToken: "page2"}, nil
		}

		return operations.PageResult[int]{}, fetchErr
	}

	_, err := operations.CollectAll(context.Background(), fetch, 0)
	assert.ErrorIs(t, err, fetchErr)
	assert.Equal(t, 2, callCount)
}
