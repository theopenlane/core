package docextract

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"google.golang.org/genai"
	"gotest.tools/v3/assert"
)

func TestSplitSectionsDropsUnderscores(t *testing.T) {
	sections, err := splitSections(`{"system_details": {"name": "demo"}, "controls": [{"refCode": "CC1.1"}]}`)

	assert.NilError(t, err)
	assert.DeepEqual(t, sections.SectionNames(), []string{"controls", "systemdetails"})
	assert.Check(t, string(sections["systemdetails"]) == `{"name": "demo"}`)
}

func TestSplitSectionsRejectsInvalidJSON(t *testing.T) {
	_, err := splitSections(`{"controls": [`)

	assert.Check(t, err != nil)
}

func TestSectionNamesEmpty(t *testing.T) {
	assert.Check(t, len(Sections{}.SectionNames()) == 0)
}

func TestCountWithoutStreamer(t *testing.T) {
	assert.Check(t, count(nil, `{"items": [1, 2, 3]}`) == 0)
}

type countingStream struct{}

func (countingStream) Merge(streamed, _ string) (string, error) { return streamed, nil }

func (countingStream) Count(output string) int {
	var payload struct {
		Items []json.RawMessage `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		return 0
	}

	return len(payload.Items)
}

func (countingStream) Continuation(context.Context, string) (string, error) { return "", nil }

func TestCountWithStreamer(t *testing.T) {
	assert.Check(t, count(countingStream{}, `{"items": [1, 2, 3]}`) == 3)
	assert.Check(t, count(countingStream{}, "") == 0)
}

func TestIsRetryableStatus(t *testing.T) {
	assert.Check(t, isRetryableStatus("CANCELLED"))
	assert.Check(t, isRetryableStatus("PARTIALLY_SUCCEEDED"))
	assert.Check(t, isRetryableStatus("UNAVAILABLE"))
	assert.Check(t, isRetryableStatus("INTERNAL"))
	assert.Check(t, isRetryableStatus("RESOURCE_EXHAUSTED"))
	assert.Check(t, !isRetryableStatus("FAILED"))
	assert.Check(t, !isRetryableStatus("INVALID_ARGUMENT"))
}

func TestExtractRequiresSystemInstruction(t *testing.T) {
	client := &Client{}

	_, err := client.Extract(context.Background(), nil, nil, Request{Section: "controls"})

	assert.ErrorIs(t, err, ErrSystemInstructionRequired)
}

type unionStream struct {
	continuation string
}

func (unionStream) Merge(streamed, collected string) (string, error) {
	merged := unionItems(collected)

	for _, item := range unionItems(streamed) {
		if !slices.Contains(merged, item) {
			merged = append(merged, item)
		}
	}

	out, err := json.Marshal(map[string][]string{"items": merged})

	return string(out), err
}

func (unionStream) Count(output string) int { return len(unionItems(output)) }

func (s unionStream) Continuation(context.Context, string) (string, error) {
	return s.continuation, nil
}

func unionItems(output string) []string {
	var payload struct {
		Items []string `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		return nil
	}

	return payload.Items
}

func TestMergePayloadContinuesWithNewItems(t *testing.T) {
	extraction := &sectionExtraction{stream: unionStream{continuation: "\nCONTINUE"}, collected: `{"items":["a"]}`}

	outcome, err := extraction.mergePayload(context.Background(), `{"items":["b"]}`, time.Now())

	assert.NilError(t, err)
	assert.Check(t, outcome == attemptContinue)
	assert.Check(t, extraction.continuation == "\nCONTINUE")
	assert.Check(t, extraction.stream.Count(extraction.collected) == 2)
}

func TestMergePayloadStopsWhenOnlyDuplicatesReturn(t *testing.T) {
	extraction := &sectionExtraction{stream: unionStream{continuation: "\nCONTINUE"}, collected: `{"items":["a"]}`}

	outcome, err := extraction.mergePayload(context.Background(), `{"items":["a"]}`, time.Now())

	assert.NilError(t, err)
	assert.Check(t, outcome == attemptDone)
	assert.Check(t, extraction.continuation == "")
}

func TestMergePayloadRetriesEmptyPayloadBeforeAnyItems(t *testing.T) {
	extraction := &sectionExtraction{stream: unionStream{}}

	outcome, err := extraction.mergePayload(context.Background(), `{"items":[]}`, time.Now())

	assert.NilError(t, err)
	assert.Check(t, outcome == attemptContinue)
	assert.Check(t, extraction.emptyAttempts == 1)
}

func TestMergePayloadStopsOnEmptyPayloadAfterItems(t *testing.T) {
	extraction := &sectionExtraction{stream: unionStream{}, collected: `{"items":["a"]}`}

	outcome, err := extraction.mergePayload(context.Background(), `{"items":[]}`, time.Now())

	assert.NilError(t, err)
	assert.Check(t, outcome == attemptDone)
}

func TestMergePayloadStopsAtMaxAttempts(t *testing.T) {
	extraction := &sectionExtraction{stream: unionStream{}, collected: `{"items":["a"]}`, attempt: maxAttempts}

	outcome, err := extraction.mergePayload(context.Background(), `{"items":["b"]}`, time.Now())

	assert.NilError(t, err)
	assert.Check(t, outcome == attemptDone)
	assert.Check(t, extraction.stream.Count(extraction.collected) == 2)
}

func TestTakeWholeOutputKeepsOutput(t *testing.T) {
	extraction := &sectionExtraction{}

	assert.Check(t, extraction.takeWholeOutput(context.Background(), `{"controls":[]}`) == attemptDone)
	assert.Check(t, extraction.collected == `{"controls":[]}`)
}

func TestTakeWholeOutputRetriesWhenEmpty(t *testing.T) {
	extraction := &sectionExtraction{}

	assert.Check(t, extraction.takeWholeOutput(context.Background(), "") == attemptContinue)
	assert.Check(t, extraction.emptyAttempts == 1)
}

func TestTakeWholeOutputStopsAfterEmptyBudget(t *testing.T) {
	extraction := &sectionExtraction{emptyAttempts: maxEmptyAttempts}

	assert.Check(t, extraction.takeWholeOutput(context.Background(), "") == attemptDone)
}

func TestResumeFailsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (&sectionExtraction{}).resume(ctx, "", generationRetry{needed: true})

	assert.ErrorIs(t, err, context.Canceled)
}

func TestResumeKeepsPartialPayloadAtMaxAttempts(t *testing.T) {
	extraction := &sectionExtraction{stream: unionStream{}, attempt: maxAttempts, collected: `{"items":["a"]}`}

	_, err := extraction.resume(context.Background(), `{"items":["b"]}`, generationRetry{needed: true})

	assert.ErrorIs(t, err, ErrMaxAttemptsReached)
	assert.Check(t, extraction.stream.Count(extraction.collected) == 2)
}

func TestRequestPartsOmitsCachedDocument(t *testing.T) {
	assert.Check(t, len(requestParts(nil, "prompt")) == 1)
	assert.Check(t, len(requestParts(&document{part: genai.NewPartFromText("doc")}, "prompt")) == 2)
}

func TestCacheIDKeepsOnlyTheTrailingIdentifier(t *testing.T) {
	assert.Check(t, CacheID("projects/123456/locations/global/cachedContents/987654") == "987654")
	assert.Check(t, CacheID("987654") == "987654")
	assert.Check(t, CacheID("") == "")
}

func TestLoadCredentialsWithoutKey(t *testing.T) {
	creds, err := LoadCredentials("")

	assert.NilError(t, err)
	assert.Check(t, creds == nil)
}

func TestLoadCredentialsRejectsMalformedKey(t *testing.T) {
	_, err := LoadCredentials("{not json")

	assert.ErrorIs(t, err, ErrCredentialsInvalid)
}
