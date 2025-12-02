//go:build windmill

package windmill

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	api "github.com/windmill-labs/windmill-go-client/api"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/ent/entconfig"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
)

var (
	errUnexpectedStatusCode = errors.New("unexpected status code")
	errEmptyResponseBody    = errors.New("empty response body")
)

const (
	defaultTimeoutSeconds = 30

	// randomIDByteLength is the number of bytes used for generating random IDs
	randomIDByteLength = 4
)

// Client is a simple SDK for creating and managing Windmill flows
type Client struct {
	baseURL         string
	workspace       string
	requester       *httpsling.Requester
	timezone        string
	onFailureScript string
	onSuccessScript string
}

// NewWindmill creates a new Windmill client based on the provided configuration
func NewWindmill(cfg entconfig.Config) (*Client, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid windmill config: %w", err)
	}

	requester, err := createRequester(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create requester: %w", err)
	}

	c := &Client{
		baseURL:         cfg.Windmill.BaseURL,
		workspace:       cfg.Windmill.Workspace,
		timezone:        cfg.Windmill.Timezone,
		onFailureScript: cfg.Windmill.OnFailureScript,
		onSuccessScript: cfg.Windmill.OnSuccessScript,
		requester:       requester,
	}

	return c, nil
}

// validateConfig validates the Windmill configuration before creating a client
func validateConfig(cfg entconfig.Config) error {
	if !cfg.Windmill.Enabled {
		return ErrWindmillDisabled
	}

	if cfg.Windmill.Token == "" {
		return ErrMissingToken
	}

	if cfg.Windmill.Workspace == "" {
		return ErrMissingWorkspace
	}

	return nil
}

// createRequester creates a new httpsling.Requester with the Windmill configuration
// It sets the base URL, authorization header, and content type for the requests
func createRequester(cfg entconfig.Config) (*httpsling.Requester, error) {
	timeout, err := time.ParseDuration(cfg.Windmill.DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse windmill default timeout: %w", err)
	}

	if timeout == 0 {
		timeout = defaultTimeoutSeconds
	}

	return httpsling.New(
		httpsling.Client(
			httpclient.Timeout(timeout),
		),

		httpsling.URL(cfg.Windmill.BaseURL),
		httpsling.Header(httpsling.HeaderAuthorization, "Bearer "+cfg.Windmill.Token),
		httpsling.Header(httpsling.HeaderContentType, httpsling.ContentTypeJSON),
	)
}

// CreateFlow creates a new flow in Windmill and returns the flow path
// It wraps raw code into proper Windmill flow structure
func (c *Client) CreateFlow(ctx context.Context, req CreateFlowRequest) (*CreateFlowResponse, error) {
	path := fmt.Sprintf("/api/w/%s/flows/create", c.workspace)

	flowValue := createFlowValue(req.Value, req.Language)

	apiReq := struct {
		Path        string `json:"path"`
		Summary     string `json:"summary,omitempty"`
		Description string `json:"description,omitempty"`
		Value       struct {
			Modules    []any `json:"modules"`
			SameWorker bool  `json:"same_worker"`
		} `json:"value"`
		Schema any `json:"schema,omitempty"`
	}{
		Path:        req.Path,
		Summary:     req.Summary,
		Description: req.Description,
		Value: struct {
			Modules    []any `json:"modules"`
			SameWorker bool  `json:"same_worker"`
		}{
			Modules:    flowValue,
			SameWorker: false,
		},
	}

	if req.Schema != nil {
		apiReq.Schema = req.Schema
	}

	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var out string

	resp, err := c.requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(path),
		httpsling.Body(jsonData),
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, fmt.Errorf("%w: %d, URL: %s, response body: %s", errUnexpectedStatusCode, resp.StatusCode, path, out)
	}

	// the response is plain text containing the flow path, not json
	flowPath := string(out)
	if flowPath == "" {
		return nil, errEmptyResponseBody
	}

	response := &CreateFlowResponse{
		Path: flowPath,
	}

	return response, nil
}

// UpdateFlow updates an existing flow in Windmill
func (c *Client) UpdateFlow(ctx context.Context, path string, req UpdateFlowRequest) error {
	url := fmt.Sprintf("%s/api/w/%s/flows/update/%s", c.baseURL, c.workspace, path)

	flowValue := createFlowValue(req.Value, req.Language)

	apiReq := struct {
		Path        string `json:"path"`
		Summary     string `json:"summary,omitempty"`
		Description string `json:"description,omitempty"`
		Value       struct {
			Modules []any `json:"modules"`
		} `json:"value"`
		Schema any `json:"schema,omitempty"`
	}{
		Path:        path,
		Summary:     req.Summary,
		Description: req.Description,
		Value: struct {
			Modules []any `json:"modules"`
		}{
			Modules: flowValue,
		},
	}

	if req.Schema != nil {
		apiReq.Schema = req.Schema
	}

	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	var out string

	resp, err := c.requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(url),
		httpsling.Body(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return fmt.Errorf("%w: %d", errUnexpectedStatusCode, resp.StatusCode)
	}

	return nil
}

// GetFlow retrieves a flow by path
func (c *Client) GetFlow(ctx context.Context, path string) (*api.Flow, error) {
	requestPath := fmt.Sprintf("api/w/%s/flows/get/%s", c.workspace, path)

	var out api.Flow

	resp, err := c.requester.ReceiveWithContext(ctx, &out,
		httpsling.Get(requestPath),
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, fmt.Errorf("%w: %d", errUnexpectedStatusCode, resp.StatusCode)
	}

	return &out, nil
}

// CreateScheduledJob creates a new scheduled job in Windmill
func (c *Client) CreateScheduledJob(ctx context.Context, req CreateScheduledJobRequest) (*CreateScheduledJobResponse, error) {
	url := fmt.Sprintf("%s/api/w/%s/schedules/create", c.baseURL, c.workspace)

	enabled := true
	apiReq := api.CreateScheduleJSONRequestBody{
		IsFlow:     true,
		Path:       req.Path,
		Schedule:   req.Schedule,
		ScriptPath: req.FlowPath,
		Enabled:    &enabled,
		Timezone:   c.timezone,
	}

	if c.onFailureScript != "" {
		apiReq.OnFailure = &c.onFailureScript
	}

	if c.onSuccessScript != "" {
		apiReq.OnSuccess = &c.onSuccessScript
	}

	if req.Args != nil {
		if args, ok := req.Args.(api.ScriptArgs); ok {
			apiReq.Args = args
		}
	}

	if req.Summary != "" {
		apiReq.Summary = &req.Summary
	}

	log.Info().Msgf("apiReq: %+v", apiReq)

	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var out string

	resp, err := c.requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(url),
		httpsling.Body(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduled job: %w", err)
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, fmt.Errorf("%w: %d, URL: %s, response body: %s", errUnexpectedStatusCode, resp.StatusCode, url, out)
	}

	// the response is plain text containing the schedule path, not json
	schedulePath := string(out)
	if schedulePath == "" {
		return nil, errEmptyResponseBody
	}

	response := &CreateScheduledJobResponse{
		Path: schedulePath,
	}

	return response, nil
}

func getWindmillLanguage(language enums.JobPlatformType) api.SchemasRawScriptLanguage {
	switch language {
	case enums.JobPlatformTypeGo:
		return api.Go
	case enums.JobPlatformTypeTs:
		return api.Bun
	default:
		// fall back to bash for any other language
		return api.Bash
	}
}

// createFlowValue creates a properly structured flow value from raw code content
func createFlowValue(rawContent []any, language enums.JobPlatformType) []any {
	modules := make([]any, 0, len(rawContent))

	for _, content := range rawContent {
		var codeContent string
		switch v := content.(type) {
		case string:
			codeContent = v
		case []byte:
			codeContent = string(v)
		default:
			if jsonBytes, err := json.Marshal(v); err == nil {
				codeContent = string(jsonBytes)
			} else {
				codeContent = fmt.Sprintf("%v", v)
			}
		}

		language := getWindmillLanguage(language)

		module := struct {
			ID    string `json:"id"`
			Value struct {
				Type            string `json:"type"`
				Content         string `json:"content"`
				Language        string `json:"language"`
				InputTransforms map[string]struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"input_transforms"`
			} `json:"value"`
		}{
			ID: generateRandomID(),
			Value: struct {
				Type            string `json:"type"`
				Content         string `json:"content"`
				Language        string `json:"language"`
				InputTransforms map[string]struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"input_transforms"`
			}{
				Type:     string(api.Rawscript),
				Content:  codeContent,
				Language: string(language),
				InputTransforms: map[string]struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				}{
					"name": {
						Type:  string(api.Static),
						Value: "",
					},
				},
			},
		}

		modules = append(modules, module)
	}

	return modules
}

// generateRandomID generates a random string ID
func generateRandomID() string {
	bytes := make([]byte, randomIDByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(bytes)
}
