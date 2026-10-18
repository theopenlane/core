package docextract

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func itemsJSON(t *testing.T, items ...map[string]any) string {
	t.Helper()

	encoded, err := json.Marshal(map[string][]map[string]any{"items": items})
	assert.NilError(t, err)

	return string(encoded)
}

func decodedItems(t *testing.T, payload string) []map[string]any {
	t.Helper()

	var section map[string][]map[string]any
	assert.NilError(t, json.Unmarshal([]byte(payload), &section))

	return section["items"]
}

func idKey(item map[string]any) (string, bool) {
	id, _ := item["id"].(string)

	return id, id != ""
}

func keepFirstTitle(previous, item map[string]any) map[string]any {
	item["title"] = previous["title"]

	return item
}

func TestSectionStreamMergeFoldsAndSorts(t *testing.T) {
	stream := &SectionStream{Section: "items", Key: idKey, SortField: "id"}

	collected := itemsJSON(t, map[string]any{"id": "b", "title": "old b"}, map[string]any{"id": "c"})
	streamed := itemsJSON(t, map[string]any{"id": "a"}, map[string]any{"id": "b", "title": "new b", "extra": true})

	merged, err := stream.Merge(streamed, collected)
	assert.NilError(t, err)

	items := decodedItems(t, merged)
	assert.Check(t, cmp.Len(items, 3))
	assert.Check(t, cmp.Equal(items[0]["id"], "a"))
	assert.Check(t, cmp.Equal(items[1]["title"], "new b"))
	assert.Check(t, cmp.Equal(items[1]["extra"], true))
	assert.Check(t, cmp.Equal(items[2]["id"], "c"))
}

func TestSectionStreamMergeAppliesFold(t *testing.T) {
	stream := &SectionStream{Section: "items", Key: idKey, Fold: keepFirstTitle}

	merged, err := stream.Merge(itemsJSON(t, map[string]any{"id": "a", "title": "second"}), itemsJSON(t, map[string]any{"id": "a", "title": "first"}))
	assert.NilError(t, err)

	items := decodedItems(t, merged)
	assert.Check(t, cmp.Len(items, 1))
	assert.Check(t, cmp.Equal(items[0]["title"], "first"))
}

func TestSectionStreamMergeKeepsUnkeyedItemsInOrder(t *testing.T) {
	stream := &SectionStream{Section: "items", Key: idKey}

	merged, err := stream.Merge(itemsJSON(t, map[string]any{"title": "x"}, map[string]any{"title": "y"}), itemsJSON(t, map[string]any{"title": "w"}))
	assert.NilError(t, err)

	items := decodedItems(t, merged)
	assert.Check(t, cmp.Len(items, 3))
	assert.Check(t, cmp.Equal(items[0]["title"], "w"))
	assert.Check(t, cmp.Equal(items[2]["title"], "y"))
}

func TestSectionStreamMergeRejectsPartialJSON(t *testing.T) {
	_, err := (&SectionStream{Section: "items", Key: idKey}).Merge(`{"items": [`, "")

	assert.Check(t, err != nil)
}

func TestSectionStreamCount(t *testing.T) {
	stream := &SectionStream{Section: "items", Key: idKey}

	assert.Check(t, stream.Count("") == 0)
	assert.Check(t, stream.Count("garbage") == 0)
	assert.Check(t, stream.Count(itemsJSON(t, map[string]any{"id": "a"}, map[string]any{"id": "b"})) == 2)
}

func TestSectionStreamContinuationPassesItems(t *testing.T) {
	var resumed []map[string]any

	stream := &SectionStream{Section: "items", Key: idKey, Resume: func(_ context.Context, items []map[string]any) (string, error) {
		resumed = items

		return "\nCONTINUE", nil
	}}

	prompt, err := stream.Continuation(context.Background(), itemsJSON(t, map[string]any{"id": "a"}))

	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(prompt, "\nCONTINUE"))
	assert.Check(t, cmp.Len(resumed, 1))
}
