package gemini

import (
	"context"
	"errors"
	"testing"
	"time"

	"net/http"

	"google.golang.org/genai"
	"google.golang.org/genproto/googleapis/rpc/code"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

var errTransport = errors.New("connection reset")

func TestIsRetryableStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status string
		want   bool
	}{
		{status: code.Code_CANCELLED.String(), want: true},
		{status: code.Code_UNAVAILABLE.String(), want: true},
		{status: code.Code_INTERNAL.String(), want: true},
		{status: code.Code_RESOURCE_EXHAUSTED.String(), want: true},
		{status: "INVALID_ARGUMENT"},
		{status: "FAILED"},
		{status: ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, isRetryableStatus(tt.status), tt.want)
		})
	}
}

func TestIsRetryableCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		responseCode int
		want         bool
	}{
		{name: "too many requests", responseCode: http.StatusTooManyRequests, want: true},
		{name: "internal server error", responseCode: http.StatusInternalServerError, want: true},
		{name: "service unavailable", responseCode: http.StatusServiceUnavailable, want: true},
		{name: "gateway timeout", responseCode: http.StatusGatewayTimeout, want: true},
		{name: "bad request", responseCode: http.StatusBadRequest},
		{name: "not found", responseCode: http.StatusNotFound},
		{name: "unset", responseCode: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, isRetryableCode(tt.responseCode), tt.want)
		})
	}
}

func TestClassify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		err                  error
		expectedRetry        bool
		expectedThrottled    bool
		expectedAfter        time.Duration
		expectedCacheMissing bool
	}{
		{
			name:              "quota rejection is throttled",
			err:               genai.APIError{Status: code.Code_RESOURCE_EXHAUSTED.String()},
			expectedRetry:     true,
			expectedThrottled: true,
		},
		{
			name:          "transient status is not throttled",
			err:           genai.APIError{Status: code.Code_UNAVAILABLE.String()},
			expectedRetry: true,
		},
		{
			name: "unretryable status",
			err:  genai.APIError{Status: "INVALID_ARGUMENT"},
		},
		{
			name:          "reason phrase body still retries on 503",
			err:           genai.APIError{Code: http.StatusServiceUnavailable, Status: "503 Service Unavailable"},
			expectedRetry: true,
		},
		{
			name:          "reason phrase body still retries on 500",
			err:           genai.APIError{Code: http.StatusInternalServerError, Status: "500 Internal Server Error"},
			expectedRetry: true,
		},
		{
			name:          "gateway timeout retries",
			err:           genai.APIError{Code: http.StatusGatewayTimeout, Status: "504 Gateway Timeout"},
			expectedRetry: true,
		},
		{
			name:              "reason phrase body still throttles on 429",
			err:               genai.APIError{Code: http.StatusTooManyRequests, Status: "429 Too Many Requests"},
			expectedRetry:     true,
			expectedThrottled: true,
		},
		{
			name: "unretryable status and code",
			err:  genai.APIError{Code: http.StatusBadRequest, Status: "INVALID_ARGUMENT"},
		},
		{
			name:                 "not found means the cache is gone",
			err:                  genai.APIError{Status: "NOT_FOUND", Code: cacheMissingCode},
			expectedCacheMissing: true,
		},
		{
			name:          "transport failure with no api error is retried",
			err:           errTransport,
			expectedRetry: true,
		},
		{
			name: "retry info supplies the delay",
			err: genai.APIError{
				Status: code.Code_RESOURCE_EXHAUSTED.String(),
				Details: []map[string]any{
					{"@type": "type.googleapis.com/google.rpc.Help"},
					{"@type": retryInfoType, "retryDelay": "42s"},
				},
			},
			expectedRetry:     true,
			expectedThrottled: true,
			expectedAfter:     42 * time.Second,
		},
		{
			name: "malformed retry info is ignored",
			err: genai.APIError{
				Status:  code.Code_RESOURCE_EXHAUSTED.String(),
				Details: []map[string]any{{"@type": retryInfoType, "retryDelay": "soon"}},
			},
			expectedRetry:     true,
			expectedThrottled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			disposition := (&Provider{}).Classify(tt.err)

			assert.Check(t, disposition.Retry == tt.expectedRetry)
			assert.Check(t, disposition.Throttled == tt.expectedThrottled)
			assert.Check(t, disposition.After == tt.expectedAfter)
			assert.Check(t, disposition.CacheMissing == tt.expectedCacheMissing)
		})
	}
}

func TestRequestParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		doc      docextract.Document
		prompt   string
		expected int
	}{
		{name: "cached document sends only the prompt", prompt: "prompt", expected: 1},
		{name: "document and prompt", doc: &document{part: genai.NewPartFromText("doc")}, prompt: "prompt", expected: 2},
		{name: "document without a prompt", doc: &document{part: genai.NewPartFromText("doc")}, expected: 1},
		{name: "neither document nor prompt"},
		{name: "document from another provider is skipped", doc: &foreignDocument{}, prompt: "prompt", expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, len(requestParts(tt.doc, tt.prompt)), tt.expected)
		})
	}
}

type foreignDocument struct{}

func (foreignDocument) Release(context.Context) {}

func TestResponseSchema(t *testing.T) {
	t.Parallel()

	nullable := true

	tests := []struct {
		name   string
		schema *docextract.Schema
		assert func(t *testing.T, converted *genai.Schema)
	}{
		{
			name: "nil schema converts to nil",
		},
		{
			name:   "string node keeps its description",
			schema: &docextract.Schema{Type: docextract.TypeString, Description: "control reference"},
			assert: func(t *testing.T, converted *genai.Schema) {
				assert.Check(t, converted.Type == genai.TypeString)
				assert.Check(t, converted.Description == "control reference")
			},
		},
		{
			name:   "boolean node",
			schema: &docextract.Schema{Type: docextract.TypeBoolean},
			assert: func(t *testing.T, converted *genai.Schema) {
				assert.Check(t, converted.Type == genai.TypeBoolean)
			},
		},
		{
			name:   "nullable and enum carry over",
			schema: &docextract.Schema{Type: docextract.TypeString, Nullable: &nullable, Enum: []string{"low", "high"}},
			assert: func(t *testing.T, converted *genai.Schema) {
				assert.Check(t, *converted.Nullable)
				assert.DeepEqual(t, converted.Enum, []string{"low", "high"})
			},
		},
		{
			name: "nested object inside an array",
			schema: &docextract.Schema{
				Type:     docextract.TypeObject,
				Required: []string{"controls"},
				Properties: map[string]*docextract.Schema{
					"controls": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type:     docextract.TypeObject,
							Required: []string{"refCode"},
							Properties: map[string]*docextract.Schema{
								"refCode": {Type: docextract.TypeString},
								"open":    {Type: docextract.TypeBoolean},
							},
						},
					},
				},
			},
			assert: func(t *testing.T, converted *genai.Schema) {
				assert.Check(t, converted.Type == genai.TypeObject)
				assert.DeepEqual(t, converted.Required, []string{"controls"})

				controls := converted.Properties["controls"]
				assert.Check(t, controls.Type == genai.TypeArray)
				assert.DeepEqual(t, controls.Items.Required, []string{"refCode"})
				assert.Check(t, controls.Items.Properties["refCode"].Type == genai.TypeString)
				assert.Check(t, controls.Items.Properties["open"].Type == genai.TypeBoolean)
			},
		},
		{
			name:   "object without properties leaves the map nil",
			schema: &docextract.Schema{Type: docextract.TypeObject},
			assert: func(t *testing.T, converted *genai.Schema) {
				assert.Check(t, converted.Properties == nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			converted := responseSchema(tt.schema)

			if tt.assert == nil {
				assert.Check(t, converted == nil)

				return
			}

			tt.assert(t, converted)
		})
	}
}

func TestNewAppliesDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		config          Config
		expectedModel   string
		expectedBackend genai.Backend
	}{
		{
			name:            "empty backend and model take the defaults",
			config:          Config{APIKey: "key"},
			expectedModel:   DefaultModel,
			expectedBackend: genai.BackendGeminiAPI,
		},
		{
			name:            "configured model is kept",
			config:          Config{APIKey: "key", Model: "gemini-custom"},
			expectedModel:   "gemini-custom",
			expectedBackend: genai.BackendGeminiAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider, err := New(context.Background(), tt.config)

			assert.NilError(t, err)
			assert.Check(t, provider.Model() == tt.expectedModel)
			assert.Check(t, provider.backend == tt.expectedBackend)
		})
	}
}
