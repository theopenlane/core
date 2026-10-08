// Package mock holds docextract test doubles for packages that build an extraction client without
// a model backend; the provider records what it was asked for so a caller can assert on it
package mock

import (
	"context"
	"errors"
	"io"
	"iter"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

// ErrProvider is the failure to inject when a test needs the provider to error
var ErrProvider = errors.New("mock: provider failed")

// Document is a prepared document that counts how often it was released
type Document struct {
	// Released is how many times Release was called
	Released int
}

// Release records the release
func (d *Document) Release(context.Context) { d.Released++ }

// Provider is a docextract.Provider that returns canned responses and records every call
type Provider struct {
	// Doc is the document PrepareDocument hands back, created on first use when nil
	Doc *Document
	// PrepareErr fails PrepareDocument when set
	PrepareErr error
	// CacheName is the cache name CreateCache returns
	CacheName string
	// CacheErr fails CreateCache when set
	CacheErr error
	// DeleteErr fails DeleteCache when set
	DeleteErr error
	// Chunks are the response payloads Stream yields in order
	Chunks []string
	// StreamErr is yielded after the chunks when set
	StreamErr error
	// Disposition is what Classify reports for every error
	Disposition docextract.Disposition

	// Prepared is how many times PrepareDocument was called
	Prepared int
	// CacheReqs are the cache requests received, in order
	CacheReqs []docextract.CacheRequest
	// Deleted are the cache names DeleteCache was called with, in order
	Deleted []string
	// Generated are the generation requests received, in order
	Generated []docextract.GenerateRequest
}

var _ docextract.Provider = (*Provider)(nil)

// PrepareDocument records the call and returns the configured document
func (p *Provider) PrepareDocument(context.Context, io.Reader) (docextract.Document, error) {
	p.Prepared++

	if p.PrepareErr != nil {
		return nil, p.PrepareErr
	}

	if p.Doc == nil {
		p.Doc = &Document{}
	}

	return p.Doc, nil
}

// CreateCache records the request and returns the configured cache name
func (p *Provider) CreateCache(_ context.Context, req docextract.CacheRequest) (string, error) {
	p.CacheReqs = append(p.CacheReqs, req)

	if p.CacheErr != nil {
		return "", p.CacheErr
	}

	return p.CacheName, nil
}

// DeleteCache records the name it was asked to delete
func (p *Provider) DeleteCache(_ context.Context, name string) error {
	p.Deleted = append(p.Deleted, name)

	return p.DeleteErr
}

// Stream records the request and yields the configured chunks, then the configured error
func (p *Provider) Stream(_ context.Context, req docextract.GenerateRequest) iter.Seq2[string, error] {
	p.Generated = append(p.Generated, req)

	return p.yieldChunks
}

// yieldChunks walks the configured chunks and ends on the configured error
func (p *Provider) yieldChunks(yield func(string, error) bool) {
	for _, chunk := range p.Chunks {
		if !yield(chunk, nil) {
			return
		}
	}

	if p.StreamErr != nil {
		yield("", p.StreamErr)
	}
}

// Classify reports the configured disposition for every error
func (p *Provider) Classify(error) docextract.Disposition { return p.Disposition }

// Kind is a docextract.Kind that returns a fixed plan for every section
type Kind struct {
	// KindName is what Name reports, defaulting to mock when empty
	KindName string
	// SectionPlan is the plan returned for every request
	SectionPlan docextract.Plan
	// Err fails Plan when set
	Err error
}

var _ docextract.Kind = Kind{}

// Name identifies the kind
func (k Kind) Name() string {
	if k.KindName == "" {
		return "mock"
	}

	return k.KindName
}

// Plan returns the configured plan
func (k Kind) Plan(docextract.Request) (docextract.Plan, error) {
	if k.Err != nil {
		return docextract.Plan{}, k.Err
	}

	return k.SectionPlan, nil
}

// NewClient builds an extraction client backed by the provider, for a caller that only needs a
// working client and does not assert on the provider
func NewClient(provider *Provider) (*docextract.Client, error) {
	return docextract.NewClient(docextract.Config{Provider: provider, SystemInstruction: "extract"})
}
