package docextract

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/pdftext"
)

const (
	maxTokens = int32(65536)
	// maxAttempts caps the continuation requests issued while merging a streamed section
	maxAttempts = 20
	// maxEmptyAttempts caps retries when the model returns no output at all
	maxEmptyAttempts = 3
	// retryDelay is the pause before re-requesting after a cancelled or partial generation
	retryDelay = 10 * time.Second
	// defaultTemperature keeps generation close to deterministic so extractions are repeatable
	defaultTemperature = 0.1
)

var temperature = float32(defaultTemperature)

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
	ResponseSchema *genai.Schema
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
		doc, err := c.loadDocument(ctx, pdf)
		if err != nil {
			return nil, err
		}

		defer c.releaseDocument(ctx, doc)

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
	doc *document
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
	contents := []*genai.Content{genai.NewContentFromParts(requestParts(e.doc, e.plan.Prompt+e.continuation), genai.RoleUser)}

	cacheName := e.cache

	var cache *genai.CachedContent

	if cacheName == "" {
		created, err := e.client.createCache(ctx, contents)
		if err != nil {
			return "", generationRetry{}, err
		}

		cache, cacheName = created, created.Name
	}

	defer e.client.deleteCache(ctx, cache)

	return e.client.streamContent(ctx, contents, e.plan.ResponseSchema, cacheName, e.stream)
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

// requestParts builds the user turn: the document when it is not already cached, then the prompt
func requestParts(doc *document, prompt string) []*genai.Part {
	var parts []*genai.Part

	if doc != nil {
		parts = append(parts, doc.part)
	}

	return append(parts, genai.NewPartFromText(prompt))
}

// count reports the streamer's item count, zero for a single-shot extraction
func count(stream Streamer, output string) int {
	if stream == nil {
		return 0
	}

	return stream.Count(output)
}

// streamContent runs one generation and returns the accumulated text; retry is true when the
// generation was cancelled or only partially succeeded and should be re-requested
func (c *Client) streamContent(ctx context.Context, contents []*genai.Content, responseSchema *genai.Schema, cache string, stream Streamer) (string, generationRetry, error) {
	iter := c.Models.GenerateContentStream(ctx, c.model, contents, c.generateConfig(responseSchema, cache))

	var buffer strings.Builder
	var merged string

	for resp, err := range iter {
		if err != nil {
			if errors.Is(err, genai.ErrPageDone) || errors.Is(err, io.EOF) {
				break
			}

			var apiErr genai.APIError
			if errors.As(err, &apiErr) {
				if retry, ok := retryFor(apiErr); ok {
					logx.FromContext(ctx).Warn().Str("status", apiErr.Status).Str("message", apiErr.Message).Str("cache", CacheID(cache)).Dur("retry_after", retry.after).Msg("docextract: retrying generation")

					return merged, retry, nil
				}

				// a shared cache that expired or was deleted is reported so the caller can rebuild it
				if cache != "" && apiErr.Code == http.StatusNotFound {
					return "", generationRetry{}, fmt.Errorf("%w: %w", ErrCacheMissing, err)
				}

				return "", generationRetry{}, fmt.Errorf("content generation error: %w", err)
			}

			// a transport failure mid-stream is retried rather than failing the whole extraction
			logx.FromContext(ctx).Warn().Err(err).Str("cache", CacheID(cache)).Msg("docextract: stream interrupted, retrying generation")

			return merged, generationRetry{needed: true}, nil
		}

		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			for _, part := range resp.Candidates[0].Content.Parts {
				if part != nil {
					buffer.WriteString(part.Text)
				}
			}
		}

		if stream == nil {
			continue
		}

		next, err := stream.Merge(buffer.String(), merged)
		if err != nil {
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

// api error statuses are canonical google api codes, not job states
const (
	statusCancelled          = "CANCELLED"
	statusPartiallySucceeded = "PARTIALLY_SUCCEEDED"
	statusUnavailable        = "UNAVAILABLE"
	statusInternal           = "INTERNAL"
	statusResourceExhausted  = "RESOURCE_EXHAUSTED"
)

// isRetryableStatus reports whether a generation error status warrants re-requesting
func isRetryableStatus(status string) bool {
	switch status {
	case statusCancelled, statusPartiallySucceeded, statusUnavailable, statusInternal, statusResourceExhausted:
		return true
	default:
		return false
	}
}

// createCache caches the document and prompt for one attempt, retrying transport failures so a
// network blip does not cost the caller a whole attempt
func (c *Client) createCache(ctx context.Context, contents []*genai.Content) (*genai.CachedContent, error) {
	config := &genai.CreateCachedContentConfig{
		Contents:          contents,
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: c.systemInstruction}}},
	}

	var lastErr error

	for attempt := range maxEmptyAttempts {
		cache, err := c.Caches.Create(ctx, c.model, config)
		if err == nil {
			return cache, nil
		}

		lastErr = err

		var apiErr genai.APIError
		if errors.As(err, &apiErr) || ctx.Err() != nil {
			break
		}

		logx.FromContext(ctx).Warn().Err(err).Int("attempt", attempt).Msg("docextract: cache creation failed, retrying")

		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to create cached content: %w", lastErr)
}

// deleteCache removes cached content created for a single attempt
func (c *Client) deleteCache(ctx context.Context, cache *genai.CachedContent) {
	if cache == nil {
		return
	}

	if _, err := c.Caches.Delete(ctx, cache.Name, nil); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("cache", CacheID(cache.Name)).Msg("docextract: failed to delete cached content")
	}
}

// document is the pdf as the model receives it: a file reference on the Gemini API, inline bytes on Vertex AI
type document struct {
	part *genai.Part
	file *genai.File
}

// loadDocument prepares the pdf for the configured backend; Vertex AI has no files api so the bytes go inline
func (c *Client) loadDocument(ctx context.Context, pdf io.Reader) (*document, error) {
	if c.backend == genai.BackendVertexAI {
		data, err := io.ReadAll(pdf)
		if err != nil {
			return nil, err
		}

		return &document{part: genai.NewPartFromBytes(data, pdftext.ContentType)}, nil
	}

	uploaded, err := c.Files.Upload(ctx, pdf, &genai.UploadFileConfig{MIMEType: pdftext.ContentType})
	if err != nil {
		return nil, err
	}

	return &document{part: genai.NewPartFromURI(uploaded.URI, uploaded.MIMEType), file: uploaded}, nil
}

// releaseDocument removes the uploaded file once extraction is finished, a no-op for inline documents
func (c *Client) releaseDocument(ctx context.Context, doc *document) {
	if doc == nil || doc.file == nil {
		return
	}

	if _, err := c.Files.Delete(ctx, doc.file.Name, nil); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("file", doc.file.Name).Msg("docextract: failed to delete uploaded file")
	}
}

// generateConfig returns the generation configuration, using the named cache when there is one
func (c *Client) generateConfig(responseSchema *genai.Schema, cache string) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{
		Temperature:      &temperature,
		MaxOutputTokens:  maxTokens,
		ResponseSchema:   responseSchema,
		ResponseMIMEType: "application/json",
	}

	if cache != "" {
		cfg.CachedContent = cache
	} else {
		cfg.SystemInstruction = &genai.Content{Parts: []*genai.Part{{Text: c.systemInstruction}}}
	}

	return cfg
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
