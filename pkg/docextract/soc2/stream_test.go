package soc2

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/docextract/schema"
)

func reviewsJSON(t *testing.T, reviews ...schema.Review) string {
	t.Helper()

	encoded, err := json.Marshal(schema.Reviews{Reviews: reviews})
	assert.NilError(t, err)

	return string(encoded)
}

func TestMergeDedupesAndSorts(t *testing.T) {
	stream := newReviewStream(testPrompts())

	collected := reviewsJSON(t, schema.Review{ExternalID: "b", Title: "old b"}, schema.Review{ExternalID: "c"})
	streamed := reviewsJSON(t, schema.Review{ExternalID: "a"}, schema.Review{ExternalID: "b", Title: "new b"})

	merged, err := stream.Merge(streamed, collected)

	assert.NilError(t, err)

	var result schema.Reviews
	assert.NilError(t, json.Unmarshal([]byte(merged), &result))
	assert.Check(t, len(result.Reviews) == 3)
	assert.Check(t, result.Reviews[0].ExternalID == "a")
	assert.Check(t, result.Reviews[1].ExternalID == "b")
	assert.Check(t, result.Reviews[1].Title == "new b")
	assert.Check(t, result.Reviews[2].ExternalID == "c")
}

func TestMergeIntoEmpty(t *testing.T) {
	stream := newReviewStream(testPrompts())

	merged, err := stream.Merge(reviewsJSON(t, schema.Review{ExternalID: "a"}), "")

	assert.NilError(t, err)
	assert.Check(t, stream.Count(merged) == 1)
}

func TestMergeRejectsPartialJSON(t *testing.T) {
	_, err := newReviewStream(testPrompts()).Merge(`{"reviews": [`, "")

	assert.Check(t, err != nil)
}

func TestReviewCount(t *testing.T) {
	assert.Check(t, ReviewCount("") == 0)
	assert.Check(t, ReviewCount("garbage") == 0)
	assert.Check(t, ReviewCount(reviewsJSON(t, schema.Review{ExternalID: "a"}, schema.Review{ExternalID: "b"})) == 2)
}

func TestContinuationTracksLastControl(t *testing.T) {
	stream := newReviewStream(testPrompts())

	first := reviewsJSON(t,
		schema.Review{ExternalID: "1", RefCodes: []string{"ACF-01", "CC1.1"}},
		schema.Review{ExternalID: "2", RefCodes: []string{"ACF-02", "CC6.1", "CC6.2"}},
	)

	prompt, err := stream.Continuation(context.Background(), first)

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nCONTINUE AFTER ACF-02 (CC6.1)")

	second := reviewsJSON(t, schema.Review{ExternalID: "3", RefCodes: []string{"ACF-01", "CC1.2"}})

	prompt, err = stream.Continuation(context.Background(), second)

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nCONTINUE AFTER ACF-01 (CC1.1)")
	assert.DeepEqual(t, stream.controls["ACF-01"], []string{"CC1.1", "CC1.2"})
}

func TestContinuationRequiresPrompt(t *testing.T) {
	prompts := testPrompts()
	prompts.Continuation = ""

	_, err := newReviewStream(prompts).Continuation(context.Background(), reviewsJSON(t, schema.Review{ExternalID: "1", RefCodes: []string{"ACF-01"}}))

	assert.ErrorIs(t, err, ErrContinuationPromptRequired)
}

func TestContinuationWithoutControls(t *testing.T) {
	prompt, err := newReviewStream(testPrompts()).Continuation(context.Background(), reviewsJSON(t, schema.Review{ExternalID: "1"}))

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nCONTINUE AFTER  ()")
}

func TestSplitRefCodes(t *testing.T) {
	control, criteria := splitRefCodes([]string{"CC6.1", "ACF-60", "CC6.2"})

	assert.Check(t, control == "ACF-60")
	assert.DeepEqual(t, criteria, []string{"CC6.1", "CC6.2"})

	control, criteria = splitRefCodes([]string{"CC6.1", "CC6.2"})

	assert.Check(t, control == "CC6.1")
	assert.DeepEqual(t, criteria, []string{"CC6.1", "CC6.2"})

	control, criteria = splitRefCodes(nil)

	assert.Check(t, control == "")
	assert.Check(t, len(criteria) == 0)
}
