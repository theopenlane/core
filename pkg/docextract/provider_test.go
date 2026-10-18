package docextract

import (
	"context"
	"errors"
	"io"
	"iter"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

var errProvider = errors.New("provider failed")

type fakeDocument struct {
	released int
}

func (d *fakeDocument) Release(context.Context) { d.released++ }

type fakeProvider struct {
	doc         *fakeDocument
	prepareErr  error
	prepared    int
	cacheName   string
	cacheErr    error
	cacheReqs   []CacheRequest
	deleted     []string
	deleteErr   error
	chunks      []string
	streamErr   error
	disposition Disposition
	generated   []GenerateRequest
}

func (p *fakeProvider) PrepareDocument(context.Context, io.Reader) (Document, error) {
	p.prepared++

	if p.prepareErr != nil {
		return nil, p.prepareErr
	}

	if p.doc == nil {
		p.doc = &fakeDocument{}
	}

	return p.doc, nil
}

func (p *fakeProvider) CreateCache(_ context.Context, req CacheRequest) (string, error) {
	p.cacheReqs = append(p.cacheReqs, req)

	if p.cacheErr != nil {
		return "", p.cacheErr
	}

	return p.cacheName, nil
}

func (p *fakeProvider) DeleteCache(_ context.Context, name string) error {
	p.deleted = append(p.deleted, name)

	return p.deleteErr
}

func (p *fakeProvider) Stream(_ context.Context, req GenerateRequest) iter.Seq2[string, error] {
	p.generated = append(p.generated, req)

	return func(yield func(string, error) bool) {
		for _, chunk := range p.chunks {
			if !yield(chunk, nil) {
				return
			}
		}

		if p.streamErr != nil {
			yield("", p.streamErr)
		}
	}
}

func (p *fakeProvider) Classify(error) Disposition { return p.disposition }

type fakeKind struct {
	plan Plan
	err  error
}

func (fakeKind) Name() string { return "fake" }

func (k fakeKind) Plan(Request) (Plan, error) { return k.plan, k.err }

type dropAllFilter struct{}

func (dropAllFilter) Filter(Sections) (Sections, int) { return Sections{}, 1 }

func TestNewClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		provider          Provider
		systemInstruction string
		expectedErr       error
	}{
		{name: "provider and instruction", provider: &fakeProvider{}, systemInstruction: "extract"},
		{name: "provider without instruction", provider: &fakeProvider{}},
		{name: "no provider", systemInstruction: "extract", expectedErr: ErrProviderRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewClient(Config{Provider: tt.provider, SystemInstruction: tt.systemInstruction})
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Check(t, client == nil)

				return
			}

			assert.NilError(t, err)
			assert.Check(t, client.provider == tt.provider)
			assert.Check(t, client.systemInstruction == tt.systemInstruction)
		})
	}
}

func TestCacheDocument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		systemInstruction string
		cacheName         string
		cacheErr          error
		prepareErr        error
		expectedName      string
		expectedErr       error
		expectedPrepared  int
		expectedReleased  int
	}{
		{
			name:              "caches and releases the upload",
			systemInstruction: "extract",
			cacheName:         "projects/1/cachedContents/9",
			expectedName:      "projects/1/cachedContents/9",
			expectedPrepared:  1,
			expectedReleased:  1,
		},
		{
			name:        "no system instruction",
			expectedErr: ErrSystemInstructionRequired,
		},
		{
			name:              "prepare fails",
			systemInstruction: "extract",
			prepareErr:        errProvider,
			expectedErr:       errProvider,
			expectedPrepared:  1,
		},
		{
			name:              "cache creation fails",
			systemInstruction: "extract",
			cacheErr:          errProvider,
			expectedErr:       errProvider,
			expectedPrepared:  1,
			expectedReleased:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc := &fakeDocument{}
			provider := &fakeProvider{doc: doc, cacheName: tt.cacheName, cacheErr: tt.cacheErr, prepareErr: tt.prepareErr}
			client := &Client{provider: provider, systemInstruction: tt.systemInstruction}

			name, err := client.CacheDocument(context.Background(), strings.NewReader("pdf"), 3*time.Hour)

			assert.Check(t, provider.prepared == tt.expectedPrepared)
			assert.Check(t, doc.released == tt.expectedReleased)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Check(t, name == "")

				return
			}

			assert.NilError(t, err)
			assert.Check(t, name == tt.expectedName)
			assert.Check(t, len(provider.cacheReqs) == 1)
			assert.Check(t, provider.cacheReqs[0].TTL == 3*time.Hour)
			assert.Check(t, provider.cacheReqs[0].SystemInstruction == tt.systemInstruction)
			assert.Check(t, provider.cacheReqs[0].Prompt == "")
		})
	}
}

func TestReleaseDocumentCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cacheName   string
		deleteErr   error
		expectedErr error
	}{
		{name: "deletes the named cache", cacheName: "cachedContents/9"},
		{name: "empty name still delegates"},
		{name: "delete fails", cacheName: "cachedContents/9", deleteErr: errProvider, expectedErr: errProvider},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider := &fakeProvider{deleteErr: tt.deleteErr}
			client := &Client{provider: provider, systemInstruction: "extract"}

			err := client.ReleaseDocumentCache(context.Background(), tt.cacheName)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NilError(t, err)
			}

			assert.DeepEqual(t, provider.deleted, []string{tt.cacheName})
		})
	}
}

func TestStreamContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		chunks          []string
		streamErr       error
		disposition     Disposition
		cache           string
		stream          Streamer
		expectedOutput  string
		expectedRetry   bool
		expectedQuota   bool
		expectedAfter   time.Duration
		expectedErr     error
		expectedWrapped bool
	}{
		{
			name:           "single chunk without streamer",
			chunks:         []string{`{"controls": []}`},
			expectedOutput: `{"controls": []}`,
		},
		{
			name:           "chunks concatenate without streamer",
			chunks:         []string{`{"cont`, `rols": []}`},
			expectedOutput: `{"controls": []}`,
		},
		{
			name:           "streamer merges each complete payload",
			chunks:         []string{`{"items":["a"]}`, `{"items":["b"]}`},
			stream:         countingStream{},
			expectedOutput: `{"items":["b"]}`,
		},
		{
			name:           "retryable failure returns what streamed",
			chunks:         []string{`{"items":["a"]}`},
			streamErr:      errProvider,
			disposition:    Disposition{Retry: true},
			stream:         countingStream{},
			expectedOutput: `{"items":["a"]}`,
			expectedRetry:  true,
		},
		{
			name:          "throttled failure carries the requested delay",
			streamErr:     errProvider,
			disposition:   Disposition{Retry: true, Throttled: true, After: 42 * time.Second},
			expectedRetry: true,
			expectedQuota: true,
			expectedAfter: 42 * time.Second,
		},
		{
			name:        "missing cache with a shared cache",
			streamErr:   errProvider,
			disposition: Disposition{CacheMissing: true},
			cache:       "cachedContents/9",
			expectedErr: ErrCacheMissing,
		},
		{
			name:            "missing cache without a shared cache",
			streamErr:       errProvider,
			disposition:     Disposition{CacheMissing: true},
			expectedErr:     errProvider,
			expectedWrapped: true,
		},
		{
			name:            "unretryable failure",
			streamErr:       errProvider,
			expectedErr:     errProvider,
			expectedWrapped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider := &fakeProvider{chunks: tt.chunks, streamErr: tt.streamErr, disposition: tt.disposition}
			client := &Client{provider: provider, systemInstruction: "extract"}

			output, retry, err := client.streamContent(context.Background(), GenerateRequest{Cache: tt.cache}, tt.stream)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Check(t, output == "")

				if tt.expectedWrapped {
					assert.Check(t, strings.Contains(err.Error(), "content generation error"))
				}

				return
			}

			assert.NilError(t, err)
			assert.Check(t, output == tt.expectedOutput)
			assert.Check(t, retry.needed == tt.expectedRetry)
			assert.Check(t, retry.quota == tt.expectedQuota)
			assert.Check(t, retry.after == tt.expectedAfter)
		})
	}
}

func TestExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		request          Request
		filter           Filterer
		chunks           []string
		expectedSections []string
		expectedPrepared int
		expectedReleased int
		expectedCaches   int
		expectedDeletes  int
	}{
		{
			name:             "prepares the document and caches per attempt",
			chunks:           []string{`{"controls": [{"refCode": "CC1.1"}], "system_details": {}}`},
			expectedSections: []string{"controls", "systemdetails"},
			expectedPrepared: 1,
			expectedReleased: 1,
			expectedCaches:   1,
			expectedDeletes:  1,
		},
		{
			name:             "shared cache skips document preparation",
			request:          Request{Section: "controls", Cache: "cachedContents/9"},
			chunks:           []string{`{"controls": [{"refCode": "CC1.1"}]}`},
			expectedSections: []string{"controls"},
		},
		{
			name:             "filter drops out of scope entries",
			request:          Request{Section: "controls", Cache: "cachedContents/9"},
			filter:           dropAllFilter{},
			chunks:           []string{`{"controls": [{"refCode": "CC9.9"}]}`},
			expectedSections: []string{},
		},
		{
			name:             "no output yields no sections",
			request:          Request{Section: "controls", Cache: "cachedContents/9"},
			expectedSections: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc := &fakeDocument{}
			provider := &fakeProvider{doc: doc, cacheName: "cachedContents/own", chunks: tt.chunks}
			client := &Client{provider: provider, systemInstruction: "extract"}

			kind := fakeKind{plan: Plan{Prompt: "list the controls", Filter: tt.filter}}

			sections, err := client.Extract(context.Background(), strings.NewReader("pdf"), kind, tt.request)

			assert.NilError(t, err)
			assert.DeepEqual(t, sections.SectionNames(), tt.expectedSections)
			assert.Check(t, provider.prepared == tt.expectedPrepared)
			assert.Check(t, doc.released == tt.expectedReleased)
			assert.Check(t, len(provider.cacheReqs) == tt.expectedCaches)
			assert.Check(t, len(provider.deleted) == tt.expectedDeletes)
		})
	}
}

func TestExtractPassesPlanToTheProvider(t *testing.T) {
	t.Parallel()

	provider := &fakeProvider{chunks: []string{`{"controls": []}`}}
	client := &Client{provider: provider, systemInstruction: "extract"}

	responseSchema := &Schema{Type: TypeObject}
	kind := fakeKind{plan: Plan{Prompt: "list the controls", ResponseSchema: responseSchema}}

	_, err := client.Extract(context.Background(), nil, kind, Request{Section: "controls", Cache: "cachedContents/9"})

	assert.NilError(t, err)
	assert.Check(t, len(provider.generated) == 1)
	assert.Check(t, provider.generated[0].Prompt == "list the controls")
	assert.Check(t, provider.generated[0].Cache == "cachedContents/9")
	assert.Check(t, provider.generated[0].ResponseSchema == responseSchema)
	assert.Check(t, provider.generated[0].SystemInstruction == "extract")
}

func TestExtractPlanError(t *testing.T) {
	t.Parallel()

	client := &Client{provider: &fakeProvider{}, systemInstruction: "extract"}

	_, err := client.Extract(context.Background(), nil, fakeKind{err: errProvider}, Request{Section: "controls"})

	assert.ErrorIs(t, err, errProvider)
}

func TestExtractPrepareDocumentError(t *testing.T) {
	t.Parallel()

	provider := &fakeProvider{prepareErr: errProvider}
	client := &Client{provider: provider, systemInstruction: "extract"}

	_, err := client.Extract(context.Background(), strings.NewReader("pdf"), fakeKind{}, Request{Section: "controls"})

	assert.ErrorIs(t, err, errProvider)
	assert.Check(t, len(provider.generated) == 0)
}

func TestExtractInvalidJSON(t *testing.T) {
	t.Parallel()

	provider := &fakeProvider{chunks: []string{`{"controls": [`}}
	client := &Client{provider: provider, systemInstruction: "extract"}

	_, err := client.Extract(context.Background(), nil, fakeKind{}, Request{Section: "controls", Cache: "cachedContents/9"})

	assert.Check(t, err != nil)
	assert.Check(t, strings.Contains(err.Error(), "failed to parse output"))
}

func TestReleaseAttemptCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		cacheName       string
		deleteErr       error
		expectedDeletes []string
	}{
		{name: "deletes the attempt cache", cacheName: "cachedContents/own", expectedDeletes: []string{"cachedContents/own"}},
		{name: "empty name is a no-op"},
		{name: "delete failure is not fatal", cacheName: "cachedContents/own", deleteErr: errProvider, expectedDeletes: []string{"cachedContents/own"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider := &fakeProvider{deleteErr: tt.deleteErr}
			client := &Client{provider: provider, systemInstruction: "extract"}

			client.releaseAttemptCache(context.Background(), tt.cacheName)

			assert.DeepEqual(t, provider.deleted, tt.expectedDeletes)
		})
	}
}
