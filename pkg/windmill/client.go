package windmill

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	wmill "github.com/windmill-labs/windmill-go-client"
	api "github.com/windmill-labs/windmill-go-client/api"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/pkg/enums"
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

type windmillRoundTripper struct {
	token string
	base  http.RoundTripper
}

// RoundTrip implements http.RoundTripper and adds authentication + content-type headers to the request
func (wrt *windmillRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())

	clonedReq.Header.Set("Authorization", "Bearer "+wrt.token)
	clonedReq.Header.Set("Content-Type", "application/json")

	return wrt.base.RoundTrip(clonedReq)
}

func (c *Client) newClient(token string) error {
	workspace := c.workspace

	client, err := api.NewClientWithResponses(c.baseURL, func(c *api.Client) error {
		c.RequestEditors = append(c.RequestEditors, func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
			return nil
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create Windmill client: %w", err)
	}

	c.wmillClient = &wmill.ClientWithWorkspace{
		Client:    client,
		Workspace: workspace,
	}

	return nil
}

// Client is a simple SDK for creating and managing Windmill flows
type Client struct {
	baseURL         string
	workspace       string
	httpClient      *http.Client
	wmillClient     *wmill.ClientWithWorkspace
	timezone        string
	onFailureScript string
	onSuccessScript string
}

// NewWindmill creates a new Windmill client based on the provided configuration
func NewWindmill(cfg entconfig.Config) (*Client, error) {
	if !cfg.Windmill.Enabled {
		return nil, ErrWindmillDisabled
	}

	if cfg.Windmill.Token == "" {
		return nil, ErrMissingToken
	}

	if cfg.Windmill.Workspace == "" {
		return nil, ErrMissingWorkspace
	}

	timeout, err := time.ParseDuration(cfg.Windmill.DefaultTimeout)
	if err != nil {
		timeout = defaultTimeoutSeconds * time.Second
	}

	transport := &windmillRoundTripper{
		token: cfg.Windmill.Token,
		base:  http.DefaultTransport,
	}

	log.Info().Msgf("Windmill client created with config: %+v", cfg.Windmill)

	c := &Client{
		baseURL:         cfg.Windmill.BaseURL,
		workspace:       cfg.Windmill.Workspace,
		timezone:        cfg.Windmill.Timezone,
		onFailureScript: cfg.Windmill.OnFailureScript,
		onSuccessScript: cfg.Windmill.OnSuccessScript,

		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}

	if err := c.newClient(cfg.Windmill.Token); err != nil {
		return nil, err
	}

	return c, nil
}

// CreateFlow creates a new flow in Windmill and returns the flow path
// It wraps raw code into proper Windmill flow structure
func (c *Client) CreateFlow(ctx context.Context, req CreateFlowRequest) (*CreateFlowResponse, error) {
	url := fmt.Sprintf("%s/api/w/%s/flows/create", c.baseURL, c.workspace)

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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d, URL: %s, response body: %s", errUnexpectedStatusCode, resp.StatusCode, url, string(respBody))
	}

	// the response is plain text containing the flow path, not json
	flowPath := string(respBody)
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
		Description: req.Summary,
		Value: struct {
			Modules []any `json:"modules"`
		}{
			Modules: flowValue,
		},
	}

	log.Info().Msgf("apiReq: %+v", apiReq)

	if req.Schema != nil {
		apiReq.Schema = req.Schema
	}

	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", errUnexpectedStatusCode, resp.StatusCode)
	}

	return nil
}

// GetFlow retrieves a flow by path
func (c *Client) GetFlow(ctx context.Context, path string) (*Flow, error) {
	url := fmt.Sprintf("%s/api/w/%s/flows/get/%s", c.baseURL, c.workspace, path)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errUnexpectedStatusCode, resp.StatusCode)
	}

	var flow Flow
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &flow, nil
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
		// TODO: use the right kind instead
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

	// resp, err := c.wmillClient.Client.CreateSchedule(ctx, c.workspace, apiReq)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create scheduled job: %w", err)
	// }

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d, URL: %s, response body: %s", errUnexpectedStatusCode, resp.StatusCode, url, string(respBody))
	}

	// the response is plain text containing the schedule path, not json
	schedulePath := string(respBody)
	if schedulePath == "" {
		return nil, errEmptyResponseBody
	}

	response := &CreateScheduledJobResponse{
		Path: schedulePath,
	}

	return response, nil
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
				Type:     "rawscript",
				Content:  codeContent,
				Language: strings.ToLower(language.String()),
				InputTransforms: map[string]struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				}{
					"name": {
						Type:  "static",
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
