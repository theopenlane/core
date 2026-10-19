package soc2

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

func sectionJSON(t *testing.T, section string, items ...map[string]any) string {
	t.Helper()

	encoded, err := json.Marshal(map[string][]map[string]any{section: items})
	assert.NilError(t, err)

	return string(encoded)
}

func reviewsJSON(t *testing.T, reviews ...map[string]any) string {
	t.Helper()

	return sectionJSON(t, ReviewsPart, reviews...)
}

func mergedItems(t *testing.T, section, merged string) []map[string]any {
	t.Helper()

	var result map[string][]map[string]any
	assert.NilError(t, json.Unmarshal([]byte(merged), &result))

	return result[section]
}

func TestMergeKeepsEveryField(t *testing.T) {
	stream := newControlStream(ReviewsPart, testPrompts()).stream()

	item := review("a", "Inspected the access list", "ACF-01", "CC6.1")
	item["summary"] = "no exceptions"
	item["approved"] = true
	item["reviewed_at"] = "2026-01-02"

	merged, err := stream.Merge(reviewsJSON(t, item), "")

	assert.NilError(t, err)

	reviews := mergedItems(t, ReviewsPart, merged)
	assert.Check(t, len(reviews) == 1)
	assert.Check(t, reviews[0]["summary"] == "no exceptions")
	assert.Check(t, reviews[0]["approved"] == true)
	assert.Check(t, reviews[0]["reviewed_at"] == "2026-01-02")
}

func TestMergeCollapsesRestatedTest(t *testing.T) {
	stream := newControlStream(ReviewsPart, testPrompts()).stream()

	details := "Inspected the written job descriptions related to a sample of job roles within Engineering"

	streamed := reviewsJSON(t,
		review("soc2-CC 1.1-01-01", details, "CC 1.1-01", "CC1.1"),
		review("soc2-CC 1.1-01-02", details, "CC 1.1-01", "CC1.5"),
	)

	merged, err := stream.Merge(streamed, "")

	assert.NilError(t, err)

	reviews := mergedItems(t, ReviewsPart, merged)
	assert.Check(t, len(reviews) == 1)
	assert.DeepEqual(t, refCodesOf(reviews[0]), []string{"CC 1.1-01", "CC1.1", "CC1.5"})
}

func TestMergeKeepsDistinctTestsForOneControl(t *testing.T) {
	stream := newControlStream(ReviewsPart, testPrompts()).stream()

	streamed := reviewsJSON(t,
		review("a", "Inspected the job descriptions", "CC 1.1-01", "CC1.1"),
		review("b", "Observed the onboarding checklist", "CC 1.1-01", "CC1.1"),
	)

	merged, err := stream.Merge(streamed, "")

	assert.NilError(t, err)
	assert.Check(t, stream.Count(merged) == 2)
}

func TestMergeCollapsesAcrossPayloads(t *testing.T) {
	stream := newControlStream(ReviewsPart, testPrompts()).stream()

	collected := reviewsJSON(t, review("a", "Inspected  the\njob descriptions", "CC 1.1-01", "CC1.1"))
	streamed := reviewsJSON(t, review("b", "inspected the job descriptions", "CC 1.1-01", "CC1.5"))

	merged, err := stream.Merge(streamed, collected)

	assert.NilError(t, err)

	reviews := mergedItems(t, ReviewsPart, merged)
	assert.Check(t, len(reviews) == 1)
	assert.DeepEqual(t, refCodesOf(reviews[0]), []string{"CC 1.1-01", "CC1.1", "CC1.5"})
}

func TestMergeKeepsReviewsWithoutTestOrID(t *testing.T) {
	stream := newControlStream(ReviewsPart, testPrompts()).stream()

	merged, err := stream.Merge(reviewsJSON(t, review("", "", "ACF-01"), review("", "", "ACF-01")), "")

	assert.NilError(t, err)
	assert.Check(t, stream.Count(merged) == 2)
}

func TestMergeCollapsesRestatedFinding(t *testing.T) {
	stream := newControlStream(FindingsPart, testPrompts()).stream()

	restated := findingWithTest("Two new hires had no background check", "Inspected background checks", "ACF-01", "CC1.4")
	again := findingWithTest("two new hires had no  background check", "Inspected background checks", "ACF-01", "CC1.5")
	other := findingWithTest("One terminated user kept access", "Inspected background checks", "ACF-01", "CC1.4")

	merged, err := stream.Merge(sectionJSON(t, FindingsPart, restated, again, other), "")

	assert.NilError(t, err)

	findings := mergedItems(t, FindingsPart, merged)
	assert.Check(t, len(findings) == 2)
	assert.DeepEqual(t, refCodesOf(findings[0]), []string{"ACF-01", "CC1.4", "CC1.5"})
}

func TestContinuationTracksLastControl(t *testing.T) {
	tracker := newControlStream(ReviewsPart, testPrompts())
	stream := tracker.stream()

	first := reviewsJSON(t,
		review("1", "", "ACF-01", "CC1.1"),
		review("2", "", "ACF-02", "CC6.1", "CC6.2"),
	)

	prompt, err := stream.Continuation(context.Background(), first)

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nCONTINUE AFTER ACF-02 (CC6.1)")

	second := reviewsJSON(t, review("3", "", "ACF-01", "CC1.2"))

	prompt, err = stream.Continuation(context.Background(), second)

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nCONTINUE AFTER ACF-01 (CC1.1)")
	assert.DeepEqual(t, tracker.controls["ACF-01"], []string{"CC1.1", "CC1.2"})
}

func TestContinuationNamesTheSection(t *testing.T) {
	prompts := testPrompts()
	prompts.Continuation = "{{ .Section }} after {{ .LastControl }}"

	prompt, err := newControlStream(FindingsPart, prompts).stream().Continuation(context.Background(), sectionJSON(t, FindingsPart, finding("x", "ACF-03")))

	assert.NilError(t, err)
	assert.Check(t, prompt == "\nfindings after ACF-03")
}

func TestContinuationRequiresPrompt(t *testing.T) {
	prompts := testPrompts()
	prompts.Continuation = ""

	_, err := newControlStream(ReviewsPart, prompts).stream().Continuation(context.Background(), reviewsJSON(t, review("1", "", "ACF-01")))

	assert.ErrorIs(t, err, ErrContinuationPromptRequired)
}

func TestContinuationWithoutControls(t *testing.T) {
	prompt, err := newControlStream(ReviewsPart, testPrompts()).stream().Continuation(context.Background(), reviewsJSON(t, review("1", "")))

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
