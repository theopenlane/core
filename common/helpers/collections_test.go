package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type foldMapAgg struct {
	sum  int
	keys map[string]struct{}
}

func TestFoldMap(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	agg := FoldMap(input, func(capacity int) foldMapAgg {
		return foldMapAgg{
			keys: make(map[string]struct{}, capacity),
		}
	}, func(acc *foldMapAgg, key string, value int) {
		acc.sum += value
		acc.keys[key] = struct{}{}
	})

	require.Equal(t, 6, agg.sum)
	require.Len(t, agg.keys, 3)
}
