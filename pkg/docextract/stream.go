package docextract

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

// SectionStream is a Streamer for any section returned as a list of items under its name, folding
// together items that share a key so a restated item lands once
type SectionStream struct {
	// Section is the name the items are listed under in each payload
	Section string
	// Key identifies an item, false when it has nothing to identify it by and is never folded
	Key func(item map[string]any) (string, bool)
	// Fold combines an item with the earlier one sharing its key; nil keeps the later item
	Fold func(previous, item map[string]any) map[string]any
	// SortField orders the merged items by that string field for stable output; empty keeps arrival order
	SortField string
	// Resume renders the prompt suffix asking the model to continue after the streamed items
	Resume func(ctx context.Context, items []map[string]any) (string, error)
}

// Merge folds the items in streamed into collected
func (s *SectionStream) Merge(streamed, collected string) (string, error) {
	incoming, err := s.items(streamed)
	if err != nil {
		return "", err
	}

	existing := []map[string]any{}

	if collected != "" {
		if existing, err = s.items(collected); err != nil {
			return "", err
		}
	}

	merged := s.fold(append(existing, incoming...))

	if s.SortField != "" {
		slices.SortStableFunc(merged, s.compare)
	}

	out, err := json.Marshal(map[string][]map[string]any{s.Section: merged})
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// Count returns how many items the output holds, zero when it is empty or malformed
func (s *SectionStream) Count(output string) int {
	if output == "" {
		return 0
	}

	items, err := s.items(output)
	if err != nil {
		return 0
	}

	return len(items)
}

// Continuation renders the prompt suffix asking the model to resume after the streamed items
func (s *SectionStream) Continuation(ctx context.Context, streamed string) (string, error) {
	items, err := s.items(streamed)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("section", s.Section).Msg("docextract: failed to read streamed items for continuation")
	}

	return s.Resume(ctx, items)
}

// items decodes the section's items from a payload
func (s *SectionStream) items(payload string) ([]map[string]any, error) {
	var section map[string][]map[string]any
	if err := json.Unmarshal([]byte(payload), &section); err != nil {
		return nil, err
	}

	return section[s.Section], nil
}

// fold collapses items sharing a key into the position of the first, keeping unkeyed items as they are
func (s *SectionStream) fold(items []map[string]any) []map[string]any {
	merged := make([]map[string]any, 0, len(items))
	position := make(map[string]int, len(items))

	for _, item := range items {
		key, ok := s.Key(item)
		if !ok {
			merged = append(merged, item)

			continue
		}

		index, seen := position[key]
		if !seen {
			position[key] = len(merged)
			merged = append(merged, item)

			continue
		}

		if s.Fold != nil {
			item = s.Fold(merged[index], item)
		}

		merged[index] = item
	}

	return merged
}

// compare orders two items by their sort field
func (s *SectionStream) compare(a, b map[string]any) int {
	return strings.Compare(mapx.Value[string](a, s.SortField), mapx.Value[string](b, s.SortField))
}
