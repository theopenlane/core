package gemini

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"time"

	"google.golang.org/genai"

	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/pdftext"
)

// DefaultModel is the Gemini model used when none is configured
const DefaultModel = "gemini-3.1-pro-preview"

const (
	// defaultTemperature keeps extraction close to the document rather than inventive
	defaultTemperature = 0.1
	// maxTokens caps one response
	maxTokens = int32(65536)
	// cacheAttempts caps retries when cache creation fails on a transport error
	cacheAttempts = 3
	// cacheRetryDelay is the pause between cache creation attempts
	cacheRetryDelay = 10 * time.Second
)

var temperature = float32(defaultTemperature)

// Config configures the Gemini provider
type Config struct {
	// APIKey authenticates with the Gemini Developer API; on Vertex AI it enables express mode
	APIKey string
	// Backend selects the api the client talks to, defaulting to the Gemini Developer API
	Backend genai.Backend
	// Project is the Google Cloud project, required for the Vertex AI backend
	Project string
	// Location is the Google Cloud region, required for the Vertex AI backend
	Location string
	// Model is the Gemini model used for content generation, defaulting to DefaultModel
	Model string
}

// Provider runs docextract generations against Gemini
type Provider struct {
	client  *genai.Client
	backend genai.Backend
	model   string
}

// ensure the provider satisfies the docextract seam
var _ docextract.Provider = (*Provider)(nil)

// New creates a Gemini provider from the supplied configuration
func New(ctx context.Context, config Config) (*Provider, error) {
	if config.Backend == genai.BackendUnspecified {
		config.Backend = genai.BackendGeminiAPI
	}

	if config.Model == "" {
		config.Model = DefaultModel
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:   config.APIKey,
		Backend:  config.Backend,
		Project:  config.Project,
		Location: config.Location,
	})
	if err != nil {
		return nil, err
	}

	return &Provider{client: client, backend: config.Backend, model: config.Model}, nil
}

// Model is the model generations run against
func (p *Provider) Model() string { return p.model }

// document is the pdf as the model receives it: a file reference on the Gemini API, inline bytes on Vertex AI
type document struct {
	client *genai.Client
	part   *genai.Part
	file   *genai.File
}

// Release removes the uploaded file once extraction is finished, a no-op for inline documents
func (d *document) Release(ctx context.Context) {
	if d == nil || d.file == nil {
		return
	}

	if _, err := d.client.Files.Delete(ctx, d.file.Name, nil); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("file", d.file.Name).Msg("docextract: failed to delete uploaded file")
	}
}

// PrepareDocument prepares the pdf for the configured backend; Vertex AI has no files api so the
// bytes go inline
func (p *Provider) PrepareDocument(ctx context.Context, pdf io.Reader) (docextract.Document, error) {
	if p.backend == genai.BackendVertexAI {
		data, err := io.ReadAll(pdf)
		if err != nil {
			return nil, err
		}

		return &document{client: p.client, part: genai.NewPartFromBytes(data, pdftext.ContentType)}, nil
	}

	uploaded, err := p.client.Files.Upload(ctx, pdf, &genai.UploadFileConfig{MIMEType: pdftext.ContentType})
	if err != nil {
		return nil, err
	}

	return &document{
		client: p.client,
		part:   genai.NewPartFromURI(uploaded.URI, uploaded.MIMEType),
		file:   uploaded,
	}, nil
}

// CreateCache stores the document, and the prompt when one is supplied, retrying transport failures
// so a network blip does not cost the caller a whole attempt
func (p *Provider) CreateCache(ctx context.Context, req docextract.CacheRequest) (string, error) {
	config := &genai.CreateCachedContentConfig{
		TTL:               req.TTL,
		Contents:          []*genai.Content{genai.NewContentFromParts(requestParts(req.Document, req.Prompt), genai.RoleUser)},
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: req.SystemInstruction}}},
	}

	var lastErr error

	for attempt := range cacheAttempts {
		cache, err := p.client.Caches.Create(ctx, p.model, config)
		if err == nil {
			return cache.Name, nil
		}

		lastErr = err

		var apiErr genai.APIError
		if errors.As(err, &apiErr) || ctx.Err() != nil {
			break
		}

		logx.FromContext(ctx).Warn().Err(err).Int("attempt", attempt).Msg("docextract: cache creation failed, retrying")

		time.Sleep(cacheRetryDelay)
	}

	return "", fmt.Errorf("failed to create cached content: %w", lastErr)
}

// DeleteCache removes cached content
func (p *Provider) DeleteCache(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}

	if _, err := p.client.Caches.Delete(ctx, name, nil); err != nil {
		return fmt.Errorf("failed to delete document cache: %w", err)
	}

	return nil
}

// Stream runs one generation, yielding the text of each response as it arrives
func (p *Provider) Stream(ctx context.Context, req docextract.GenerateRequest) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		contents := []*genai.Content{genai.NewContentFromParts(requestParts(req.Document, req.Prompt), genai.RoleUser)}

		for resp, err := range p.client.Models.GenerateContentStream(ctx, p.model, contents, p.generateConfig(req)) {
			if err != nil {
				if errors.Is(err, genai.ErrPageDone) || errors.Is(err, io.EOF) {
					return
				}

				yield("", err)

				return
			}

			if !yield(responseText(resp), nil) {
				return
			}
		}
	}
}

// Classify reports whether a generation error warrants another attempt; a transport failure with no
// api error attached is a mid-stream interruption and is always worth re-requesting
func (p *Provider) Classify(err error) docextract.Disposition {
	var apiErr genai.APIError
	if !errors.As(err, &apiErr) {
		return docextract.Disposition{Retry: true}
	}

	if isRetryableStatus(apiErr.Status) || isRetryableCode(apiErr.Code) {
		return docextract.Disposition{
			Retry:     true,
			Throttled: isThrottled(apiErr.Status, apiErr.Code),
			After:     retryInfoDelay(apiErr.Details),
		}
	}

	return docextract.Disposition{CacheMissing: apiErr.Code == cacheMissingCode}
}

// responseText is the text of every part of a response's first candidate
func responseText(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return ""
	}

	var text string

	for _, part := range resp.Candidates[0].Content.Parts {
		if part != nil {
			text += part.Text
		}
	}

	return text
}

// requestParts builds the user turn: the document when it is not already cached, then the prompt
func requestParts(doc docextract.Document, prompt string) []*genai.Part {
	var parts []*genai.Part

	if geminiDoc, ok := doc.(*document); ok && geminiDoc != nil && geminiDoc.part != nil {
		parts = append(parts, geminiDoc.part)
	}

	if prompt == "" {
		return parts
	}

	return append(parts, genai.NewPartFromText(prompt))
}

// generateConfig returns the generation configuration, using the named cache when there is one
func (p *Provider) generateConfig(req docextract.GenerateRequest) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{
		Temperature:      &temperature,
		MaxOutputTokens:  maxTokens,
		ResponseSchema:   responseSchema(req.ResponseSchema),
		ResponseMIMEType: "application/json",
	}

	if req.Cache != "" {
		cfg.CachedContent = req.Cache
	} else {
		cfg.SystemInstruction = &genai.Content{Parts: []*genai.Part{{Text: req.SystemInstruction}}}
	}

	return cfg
}
