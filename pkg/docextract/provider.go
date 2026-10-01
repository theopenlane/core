package docextract

import (
	"context"
	"io"
	"iter"
	"time"
)

// Document is a pdf prepared for the model, released once the extraction that prepared it finishes
type Document interface {
	// Release removes whatever provider side resource the document holds
	Release(ctx context.Context)
}

// CacheRequest asks a provider to hold a document, and optionally a prompt, for later generations
type CacheRequest struct {
	// Document is the prepared pdf to cache
	Document Document
	// Prompt is cached with the document when set, so a repeated generation resends neither
	Prompt string
	// SystemInstruction is the system prompt stored alongside the cache
	SystemInstruction string
	// TTL is how long the cache lives; zero leaves the provider default
	TTL time.Duration
}

// GenerateRequest is one generation: what to send the model and what shape to return
type GenerateRequest struct {
	// Document is the prepared pdf, nil when Cache already holds it
	Document Document
	// Prompt is the full prompt sent alongside the document
	Prompt string
	// ResponseSchema constrains the model output
	ResponseSchema *Schema
	// Cache names cached content holding the document, empty when the document travels with the request
	Cache string
	// SystemInstruction is the system prompt, sent only when no cache already holds it
	SystemInstruction string
}

// Disposition classifies a generation failure so the extraction loop knows how to proceed
type Disposition struct {
	// Retry is true when the generation should be re-requested
	Retry bool
	// Throttled is true when the request was rejected for exceeding a quota, so concurrent
	// extractions must not all come back at the same moment
	Throttled bool
	// After is the delay the provider asked for, zero when it supplied none
	After time.Duration
	// CacheMissing is true when the cache named in the request no longer exists
	CacheMissing bool
}

// Provider is the model backend an extraction runs against; the gemini subpackage is one
// implementation. It owns everything sdk specific: uploading the document, caching it, running a
// generation, and deciding whether a failure is worth another attempt
type Provider interface {
	// PrepareDocument makes the pdf available to the model for the extraction that follows
	PrepareDocument(ctx context.Context, pdf io.Reader) (Document, error)
	// CreateCache stores a document, and optionally a prompt, returning the cache name
	CreateCache(ctx context.Context, req CacheRequest) (string, error)
	// DeleteCache removes cached content; a cache that already expired is not an error
	DeleteCache(ctx context.Context, name string) error
	// Stream runs one generation, yielding the text of each response as it arrives and stopping
	// at the end of the stream
	Stream(ctx context.Context, req GenerateRequest) iter.Seq2[string, error]
	// Classify says whether a generation or cache error warrants another attempt
	Classify(err error) Disposition
}
