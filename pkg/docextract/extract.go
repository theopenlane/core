package docextract

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// maxAttempts caps the continuation requests issued while merging a streamed section
	maxAttempts = 20
	// maxEmptyAttempts caps retries when the model returns no output at all
	maxEmptyAttempts = 3
	// retryDelay is the pause before re-requesting after a cancelled or partial generation
	retryDelay = 10 * time.Second
)

// Request selects what to extract from a document
type Request struct {
	// Section names the part of the document to extract, as the Kind understands it
	Section string
	// Scope optionally limits the extraction to the listed identifiers, e.g. control ref codes
	Scope []string
	// Cache names cached content already holding the document, so the pdf is not re-sent per request
	Cache string
}

// Plan is what a Kind produces for a request: the prompt, the response schema the model must
// follow, and optionally a Streamer when the section arrives as many payloads
type Plan struct {
	// Prompt is the full prompt sent alongside the document
	Prompt string
	// ResponseSchema constrains the model output
	ResponseSchema *Schema
	// Stream, when set, merges streamed payloads and continues output the model trimmed
	Stream Streamer
	// Filter, when set, drops extracted entries that fall outside the request, e.g. fabricated items
	Filter Filterer
}

// Filterer prunes extracted sections after the model output is parsed
type Filterer interface {
	// Filter returns the sections with out-of-scope entries removed and how many were dropped
	Filter(sections Sections) (Sections, int)
}

// Kind describes one kind of document the client can extract from
type Kind interface {
	// Name identifies the kind, e.g. soc2_report
	Name() string
	// Plan builds the prompt and schema for one request
	Plan(req Request) (Plan, error)
}

// Streamer handles a section the model returns as a series of payloads that must be merged,
// continuing from where the previous response stopped until nothing new arrives
type Streamer interface {
	// Merge folds a streamed payload into the collected output
	Merge(streamed, collected string) (string, error)
	// Count reports how many items the output holds, zero when it is empty or malformed
	Count(output string) int
	// Continuation returns the prompt suffix asking the model to resume after the streamed output
	Continuation(ctx context.Context, streamed string) (string, error)
}

// Sections is the extracted document keyed by section name, e.g. controls, reviews, entities
type Sections map[string]json.RawMessage

// SectionNames lists the extracted section names in sorted order
func (s Sections) SectionNames() []string {
	names := make([]string, 0, len(s))
	for name := range s {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// Extract uploads the pdf, prompts the model for the requested section of the given kind, and
// returns the extracted sections keyed by name
func (c *Client) Extract(ctx context.Context, pdf io.Reader, kind Kind, req Request) (Sections, error) {
	if c.systemInstruction == "" {
		return nil, ErrSystemInstructionRequired
	}

	plan, err := kind.Plan(req)
	if err != nil {
		return nil, err
	}

	ctx = logx.WithFields(ctx, logx.LogFields{"kind": kind.Name(), "section": req.Section})

	extraction := &sectionExtraction{client: c, plan: plan, cache: req.Cache, stream: plan.Stream, started: time.Now()}

	// a shared cache already holds the document, so it is only loaded when there is none
	if req.Cache == "" {
		doc, err := c.provider.PrepareDocument(ctx, pdf)
		if err != nil {
			return nil, err
		}

		defer doc.Release(ctx)

		extraction.doc = doc
	}

	if err := extraction.collect(ctx); err != nil {
		return nil, err
	}

	return extraction.sections(ctx)
}

// sectionExtraction is one Extract call in progress: what to ask for, and what has been collected
// across the attempts made so far
type sectionExtraction struct {
	// client issues the generations and owns the per-attempt cache
	client *Client
	// plan is the prompt, schema, streamer, and filter for the requested section
	plan Plan
	// doc is the document sent with every request, nil when a shared cache already holds it
	doc Document
	// cache names shared cached content holding the document, empty when each attempt caches its own
	cache string
	// stream merges payloads for a section the model returns in pieces, nil for a single-shot section
	stream Streamer

	// attempt counts the generations issued, including continuations and retries
	attempt int
	// emptyAttempts counts consecutive generations that returned nothing usable
	emptyAttempts int
	// continuation is the prompt suffix asking the model to resume, empty on the first attempt
	continuation string
	// collected is the merged output so far
	collected string
	// started is when the first generation was issued, for elapsed timing
	started time.Time
}

// attemptOutcome says whether the loop should issue another generation
type attemptOutcome int

const (
	// attemptContinue asks for another generation, either a retry or a continuation
	attemptContinue attemptOutcome = iota
	// attemptDone means nothing more can be collected
	attemptDone
)

// collect runs generations until the section is complete, the model stops adding items, or the
// attempt budget runs out
func (e *sectionExtraction) collect(ctx context.Context) error {
	for {
		logx.FromContext(ctx).Debug().Int("attempt", e.attempt).Int("max_attempts", maxAttempts).Int("collected_items", count(e.stream, e.collected)).Int("collected_bytes", len(e.collected)).Dur("elapsed", time.Since(e.started)).Bool("continuation", e.continuation != "").Bool("shared_cache", e.cache != "").Msg("docextract: starting request")

		attemptStarted := time.Now()

		output, retry, err := e.generate(ctx)
		if err != nil {
			return err
		}

		var outcome attemptOutcome

		switch {
		case retry.needed:
			outcome, err = e.resume(ctx, output, retry)
		case e.stream == nil:
			outcome = e.takeWholeOutput(ctx, output)
		default:
			outcome, err = e.mergePayload(ctx, output, attemptStarted)
		}

		if err != nil {
			return err
		}

		if outcome == attemptDone {
			return nil
		}

		e.attempt++
	}
}

// generate runs one generation, caching the document and prompt for this attempt alone when no
// shared cache was supplied
func (e *sectionExtraction) generate(ctx context.Context) (string, generationRetry, error) {
	prompt := e.plan.Prompt + e.continuation

	cacheName := e.cache
	ownCache := ""

	if cacheName == "" {
		created, err := e.client.provider.CreateCache(ctx, CacheRequest{
			Document:          e.doc,
			Prompt:            prompt,
			SystemInstruction: e.client.systemInstruction,
		})
		if err != nil {
			return "", generationRetry{}, err
		}

		cacheName, ownCache = created, created
	}

	defer e.client.releaseAttemptCache(ctx, ownCache)

	return e.client.streamContent(ctx, GenerateRequest{
		Document:          e.doc,
		Prompt:            prompt,
		ResponseSchema:    e.plan.ResponseSchema,
		Cache:             cacheName,
		SystemInstruction: e.client.systemInstruction,
	}, e.stream)
}

// resume keeps whatever complete payloads streamed before an interruption and waits out the retry
// delay before the next generation
func (e *sectionExtraction) resume(ctx context.Context, output string, retry generationRetry) (attemptOutcome, error) {
	if ctx.Err() != nil {
		return attemptDone, fmt.Errorf("content generation error: %w", ctx.Err())
	}

	if e.stream != nil && output != "" {
		if merged, err := e.stream.Merge(output, e.collected); err == nil {
			e.collected = merged
		}
	}

	if e.attempt >= maxAttempts {
		return attemptDone, fmt.Errorf("content generation error: %w", ErrMaxAttemptsReached)
	}

	wait := retry.wait(e.attempt)

	logx.FromContext(ctx).Debug().Int("attempt", e.attempt).Dur("wait", wait).Bool("quota", retry.quota).Msg("docextract: waiting before the next generation")

	time.Sleep(wait)

	return attemptContinue, nil
}

// takeWholeOutput accepts the output of a section that arrives in one piece, retrying only while
// the model returns nothing at all
func (e *sectionExtraction) takeWholeOutput(ctx context.Context, output string) attemptOutcome {
	e.collected = output

	if output == "" && e.emptyAttempts < maxEmptyAttempts {
		e.emptyAttempts++

		logx.FromContext(ctx).Warn().Int("attempt", e.attempt).Msg("docextract: no output generated, retrying")

		return attemptContinue
	}

	return attemptDone
}

// mergePayload folds one streamed payload into the collected output and decides whether the
// document still has items left to return
func (e *sectionExtraction) mergePayload(ctx context.Context, output string, attemptStarted time.Time) (attemptOutcome, error) {
	streamed := e.stream.Count(output)
	before := e.stream.Count(e.collected)

	if streamed == 0 {
		e.emptyAttempts++

		if before == 0 && e.emptyAttempts < maxEmptyAttempts {
			logx.FromContext(ctx).Warn().Int("attempt", e.attempt).Msg("docextract: no items generated, retrying")

			return attemptContinue, nil
		}

		return attemptDone, nil
	}

	merged, err := e.stream.Merge(output, e.collected)
	if err != nil {
		return attemptDone, fmt.Errorf("failed to merge streamed output: %w", err)
	}

	e.collected = merged

	after := e.stream.Count(e.collected)

	logx.FromContext(ctx).Debug().Int("attempt", e.attempt).Dur("attempt_duration", time.Since(attemptStarted)).Int("streamed_items", streamed).Int("new_items", after-before).Int("duplicate_items", streamed-(after-before)).Int("collected_items", after).Msg("docextract: merged streamed output")

	// a continuation that only re-sends items already collected means the document is exhausted
	if after == before {
		logx.FromContext(ctx).Debug().Int("collected_items", after).Int("attempts", e.attempt+1).Dur("elapsed", time.Since(e.started)).Msg("docextract: no new items returned, section complete")

		return attemptDone, nil
	}

	if e.attempt >= maxAttempts {
		logx.FromContext(ctx).Warn().Int("attempt", e.attempt).Int("collected_items", after).Dur("elapsed", time.Since(e.started)).Msg("docextract: reached max retries, treating output as complete")

		return attemptDone, nil
	}

	e.continuation, err = e.stream.Continuation(ctx, output)
	if err != nil {
		return attemptDone, err
	}

	return attemptContinue, nil
}

// sections splits the collected output into one payload per section, dropping entries the plan
// considers out of scope
func (e *sectionExtraction) sections(ctx context.Context) (Sections, error) {
	if e.collected == "" {
		logx.FromContext(ctx).Warn().Int("attempts", e.attempt).Msg("docextract: no output generated, returning empty result")

		return Sections{}, nil
	}

	sections, err := splitSections(e.collected)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("docextract: model output is not valid json")

		return nil, fmt.Errorf("failed to parse output after %d attempts: %w", e.attempt, err)
	}

	if e.plan.Filter != nil {
		var dropped int

		sections, dropped = e.plan.Filter.Filter(sections)
		if dropped > 0 {
			logx.FromContext(ctx).Warn().Int("dropped_items", dropped).Msg("docextract: dropped entries outside the requested scope")
		}
	}

	return sections, nil
}

// count reports the streamer's item count, zero for a single-shot extraction
func count(stream Streamer, output string) int {
	if stream == nil {
		return 0
	}

	return stream.Count(output)
}

// streamContent runs one generation and returns the accumulated text; retry is true when the
// provider reports the generation should be re-requested
func (c *Client) streamContent(ctx context.Context, req GenerateRequest, stream Streamer) (string, generationRetry, error) {
	var buffer strings.Builder
	var merged string

	for chunk, err := range c.provider.Stream(ctx, req) {
		if err != nil {
			disposition := c.provider.Classify(err)

			if disposition.Retry {
				retry := generationRetry{needed: true, quota: disposition.Throttled, after: disposition.After}

				logx.FromContext(ctx).Warn().Err(err).Str("cache", CacheID(req.Cache)).Dur("retry_after", retry.after).Msg("docextract: retrying generation")

				return merged, retry, nil
			}

			// a shared cache that expired or was deleted is reported so the caller can rebuild it
			if disposition.CacheMissing && req.Cache != "" {
				return "", generationRetry{}, fmt.Errorf("%w: %w", ErrCacheMissing, err)
			}

			return "", generationRetry{}, fmt.Errorf("content generation error: %w", err)
		}

		buffer.WriteString(chunk)

		if stream == nil {
			continue
		}

		next, mergeErr := stream.Merge(buffer.String(), merged)
		if mergeErr != nil {
			continue
		}

		// merged cleanly, so reset the buffer for the next streamed payload
		merged = next

		buffer.Reset()
	}

	if stream == nil {
		return buffer.String(), generationRetry{}, nil
	}

	return merged, generationRetry{}, nil
}

// releaseAttemptCache removes cached content created for a single attempt
func (c *Client) releaseAttemptCache(ctx context.Context, name string) {
	if name == "" {
		return
	}

	if err := c.provider.DeleteCache(ctx, name); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("cache", CacheID(name)).Msg("docextract: failed to delete cached content")
	}
}

// splitSections splits the top-level json object into one raw payload per section, dropping
// underscores from keys so system_details becomes systemdetails
func splitSections(resp string) (Sections, error) {
	var result map[string]json.RawMessage

	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, err
	}

	sections := make(Sections, len(result))

	for key, value := range result {
		sections[strings.ReplaceAll(key, "_", "")] = value
	}

	return sections, nil
}
