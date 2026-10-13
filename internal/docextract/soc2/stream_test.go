package soc2

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

func reviewsJSON(t *testing.T, reviews ...reviewEntry) string {
	t.Helper()

	encoded, err := json.Marshal(reviewSection{Reviews: reviews})
	assert.NilError(t, err)

	return string(encoded)
}

func TestMergeDedupesAndSorts(t *testing.T) {
	stream := newReviewStream(testPrompts())

	collected := reviewsJSON(t, reviewEntry{ExternalID: "b", Title: "old b"}, reviewEntry{ExternalID: "c"})
	streamed := reviewsJSON(t, reviewEntry{ExternalID: "a"}, reviewEntry{ExternalID: "b", Title: "new b"})

	merged, err := stream.Merge(streamed, collected)

	assert.NilError(t, err)

	var result reviewSection
	assert.NilError(t, json.Unmarshal([]byte(merged), &result))
	assert.Check(t, len(result.Reviews) == 3)
	assert.Check(t, result.Reviews[0].ExternalID == "a")
	assert.Check(t, result.Reviews[1].ExternalID == "b")
	assert.Check(t, result.Reviews[1].Title == "new b")
	assert.Check(t, result.Reviews[2].ExternalID == "c")
}

func TestMergeCollapsesRestatedTest(t *testing.T) {
	stream := newReviewStream(testPrompts())

	details := "Inspected the written job descriptions related to a sample of job roles within Engineering"

	streamed := reviewsJSON(t,
		reviewEntry{ExternalID: "soc2-CC 1.1-01-01", Details: details, RefCodes: []string{"CC 1.1-01", "CC1.1"}},
		reviewEntry{ExternalID: "soc2-CC 1.1-01-02", Details: details, RefCodes: []string{"CC 1.1-01", "CC1.5"}},
	)

	merged, err := stream.Merge(streamed, "")

	assert.NilError(t, err)

	var result reviewSection
	assert.NilError(t, json.Unmarshal([]byte(merged), &result))
	assert.Check(t, len(result.Reviews) == 1)
	assert.DeepEqual(t, result.Reviews[0].RefCodes, []string{"CC 1.1-01", "CC1.1", "CC1.5"})
}

func TestMergeKeepsDistinctTestsForOneControl(t *testing.T) {
	stream := newReviewStream(testPrompts())

	streamed := reviewsJSON(t,
		reviewEntry{ExternalID: "a", Details: "Inspected the job descriptions", RefCodes: []string{"CC 1.1-01", "CC1.1"}},
		reviewEntry{ExternalID: "b", Details: "Observed the onboarding checklist", RefCodes: []string{"CC 1.1-01", "CC1.1"}},
	)

	merged, err := stream.Merge(streamed, "")

	assert.NilError(t, err)
	assert.Check(t, stream.Count(merged) == 2)
}

func TestMergeCollapsesAcrossPayloads(t *testing.T) {
	stream := newReviewStream(testPrompts())

	collected := reviewsJSON(t, reviewEntry{ExternalID: "a", Details: "Inspected  the\njob descriptions", RefCodes: []string{"CC 1.1-01", "CC1.1"}})
	streamed := reviewsJSON(t, reviewEntry{ExternalID: "b", Details: "inspected the job descriptions", RefCodes: []string{"CC 1.1-01", "CC1.5"}})

	merged, err := stream.Merge(streamed, collected)

	assert.NilError(t, err)

	var result reviewSection
	assert.NilError(t, json.Unmarshal([]byte(merged), &result))
	assert.Check(t, len(result.Reviews) == 1)
	assert.DeepEqual(t, result.Reviews[0].RefCodes, []string{"CC 1.1-01", "CC1.1", "CC1.5"})
}

func TestMergeIntoEmpty(t *testing.T) {
	stream := newReviewStream(testPrompts())

	merged, err := stream.Merge(reviewsJSON(t, reviewEntry{ExternalID: "a"}), "")

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
	assert.Check(t, ReviewCount(reviewsJSON(t, reviewEntry{ExternalID: "a"}, reviewEntry{ExternalID: "b"})) == 2)
}

func TestContinuationTracksLastControl(t *testing.T) {
	stream := newReviewStream(testPrompts())

	first := reviewsJSON(t,
		reviewEntry{ExternalID: "1", RefCodes: []string{"ACF-01", "CC1.1"}},
		reviewEntry{ExternalID: "2", RefCodes: []string{"ACF-02", "CC6.1", "CC6.2"}},
	)

	prompt, err := stream.Continuation(context.Background(), first)

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nCONTINUE AFTER ACF-02 (CC6.1)")

	second := reviewsJSON(t, reviewEntry{ExternalID: "3", RefCodes: []string{"ACF-01", "CC1.2"}})

	prompt, err = stream.Continuation(context.Background(), second)

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nCONTINUE AFTER ACF-01 (CC1.1)")
	assert.DeepEqual(t, stream.controls["ACF-01"], []string{"CC1.1", "CC1.2"})
}

func TestContinuationRequiresPrompt(t *testing.T) {
	prompts := testPrompts()
	prompts.Continuation = ""

	_, err := newReviewStream(prompts).Continuation(context.Background(), reviewsJSON(t, reviewEntry{ExternalID: "1", RefCodes: []string{"ACF-01"}}))

	assert.ErrorIs(t, err, ErrContinuationPromptRequired)
}

func TestContinuationWithoutControls(t *testing.T) {
	prompt, err := newReviewStream(testPrompts()).Continuation(context.Background(), reviewsJSON(t, reviewEntry{ExternalID: "1"}))

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
