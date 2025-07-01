package windmill

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/theopenlane/core/internal/ent/entconfig"
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

// Client is a simple SDK for creating and managing Windmill flows
type Client struct {
	baseURL         string
	workspace       string
	httpClient      *http.Client
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
		timeout = 30 * time.Second
	}

	transport := &windmillRoundTripper{
		token: cfg.Windmill.Token,
		base:  http.DefaultTransport,
	}

	return &Client{
		baseURL:         cfg.Windmill.BaseURL,
		workspace:       cfg.Windmill.Workspace,
		timezone:        cfg.Windmill.Timezone,
		onFailureScript: cfg.Windmill.OnFailureScript,
		onSuccessScript: cfg.Windmill.OnSuccessScript,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

// CreateFlow creates a new flow in Windmill and returns the flow path
// It also appends a default notification flow
func (c *Client) CreateFlow(ctx context.Context, req CreateFlowRequest) (*CreateFlowResponse, error) {
	url := fmt.Sprintf("%s/api/w/%s/flows/create", c.baseURL, c.workspace)

	modules := appendWindmillModule(req.Value)

	apiReq := struct {
		Path    string `json:"path"`
		Summary string `json:"summary"`
		Value   struct {
			Modules []any `json:"modules"`
		} `json:"value"`
		Schema any `json:"schema,omitempty"`
	}{
		Path:    req.Path,
		Summary: req.Summary,
		Value: struct {
			Modules []any `json:"modules"`
		}{
			Modules: modules,
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
		return nil, fmt.Errorf("unexpected status code: %d, URL: %s, response body: %s", resp.StatusCode, url, string(respBody))
	}

	// the response is plain text containing the flow path, not JSON
	flowPath := string(respBody)
	if flowPath == "" {
		return nil, fmt.Errorf("empty response body")
	}

	response := &CreateFlowResponse{
		Path: flowPath,
	}

	return response, nil
}

// UpdateFlow updates an existing flow in Windmill
func (c *Client) UpdateFlow(ctx context.Context, path string, req UpdateFlowRequest) error {
	url := fmt.Sprintf("%s/api/w/%s/flows/update/%s", c.baseURL, c.workspace, path)

	modules := appendWindmillModule(req.Value)

	apiReq := struct {
		Summary string `json:"summary"`
		Value   struct {
			Path    string `json:"path"`
			Summary string `json:"summary"`
			Value   struct {
				Modules []any `json:"modules"`
			} `json:"value"`
		} `json:"value"`
		Schema any `json:"schema,omitempty"`
	}{
		Summary: req.Summary,
		Value: struct {
			Path    string `json:"path"`
			Summary string `json:"summary"`
			Value   struct {
				Modules []any `json:"modules"`
			} `json:"value"`
		}{
			Path:    path,
			Summary: req.Summary,
			Value: struct {
				Modules []any `json:"modules"`
			}{
				Modules: modules,
			},
		},
	}

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
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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

	apiReq := struct {
		Path       string `json:"path"`
		Schedule   string `json:"schedule"`
		ScriptPath string `json:"script_path"`
		Enabled    bool   `json:"enabled"`
		Timezone   string `json:"timezone,omitempty"`
		OnFailure  *struct {
			ScriptPath string `json:"script_path"`
			Args       any    `json:"args,omitempty"`
		} `json:"on_failure,omitempty"`
		OnSuccess *struct {
			ScriptPath string `json:"script_path"`
			Args       any    `json:"args,omitempty"`
		} `json:"on_success,omitempty"`
		Args    any    `json:"args,omitempty"`
		Summary string `json:"summary,omitempty"`
	}{
		Path:       req.Path,
		Schedule:   req.Schedule,
		ScriptPath: req.FlowPath,
		Enabled:    true,
		Timezone:   c.timezone,
	}

	if c.onFailureScript != "" {
		apiReq.OnFailure = &struct {
			ScriptPath string `json:"script_path"`
			Args       any    `json:"args,omitempty"`
		}{
			ScriptPath: c.onFailureScript,
			Args:       req.Args,
		}
	}

	if c.onSuccessScript != "" {
		apiReq.OnSuccess = &struct {
			ScriptPath string `json:"script_path"`
			Args       any    `json:"args,omitempty"`
		}{
			ScriptPath: c.onSuccessScript,
			Args:       req.Args,
		}
	}

	if req.Args != nil {
		apiReq.Args = req.Args
	}

	if req.Summary != "" {
		apiReq.Summary = req.Summary
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
		return nil, fmt.Errorf("unexpected status code: %d, URL: %s, response body: %s", resp.StatusCode, url, string(respBody))
	}

	// The response is plain text containing the schedule path, not JSON
	schedulePath := string(respBody)
	if schedulePath == "" {
		return nil, fmt.Errorf("empty response body")
	}

	response := &CreateScheduledJobResponse{
		Path: schedulePath,
	}

	return response, nil
}

// appendWindmillModule appends the standard Windmill Go module to the modules array
func appendWindmillModule(modules []any) []any {
	windmillModule := struct {
		ID    string `json:"id"`
		Value struct {
			Lock            string `json:"lock"`
			Type            string `json:"type"`
			Content         string `json:"content"`
			Language        string `json:"language"`
			InputTransforms struct {
				X struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"x"`
				Nested struct {
					Type  string `json:"type"`
					Value struct {
						Foo string `json:"foo"`
					} `json:"value"`
				} `json:"nested"`
			} `json:"input_transforms"`
		} `json:"value"`
	}{
		ID: generateRandomID(),
		Value: struct {
			Lock            string `json:"lock"`
			Type            string `json:"type"`
			Content         string `json:"content"`
			Language        string `json:"language"`
			InputTransforms struct {
				X struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"x"`
				Nested struct {
					Type  string `json:"type"`
					Value struct {
						Foo string `json:"foo"`
					} `json:"value"`
				} `json:"nested"`
			} `json:"input_transforms"`
		}{
			Lock: `module mymod

go 1.22.5

require github.com/windmill-labs/windmill-go-client v1.501.4

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
)

//go.sum

github.com/RaveNoX/go-jsoncommentstrip v1.0.0/go.mod h1:78ihd09MekBnJnxpICcwzCMzGrKSKYe4AqU6PDYYpjk=
github.com/apapsch/go-jsonmerge/v2 v2.0.0 h1:axGnT1gRIfimI7gJifB699GoE/oq+F2MU7Dml6nw9rQ=
github.com/apapsch/go-jsonmerge/v2 v2.0.0/go.mod h1:lvDnEdqiQrp0O42VQGgmlKpxL1AP2+08jFMw88y4klk=
github.com/bmatcuk/doublestar v1.1.1/go.mod h1:UD6OnuiIn0yFxxA2le/rnRU1G4RaI4UvFv1sNto9p6w=
github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/davecgh/go-spew v1.1.1 h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=
github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/google/uuid v1.5.0 h1:1p67kYwdtXjb0gL0BPiP1Av9wiZPo5A8z2cWkTZ+eyU=
github.com/google/uuid v1.5.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d/go.mod h1:2PavIy+JPciBPrBUjwbNvtwB6RQlve+hkpll6QSNmOE=
github.com/oapi-codegen/runtime v1.1.1 h1:EXLHh0DXIJnWhdRPN2w4MXAzFyE4CskzhNLUmtpMYro=
github.com/oapi-codegen/runtime v1.1.1/go.mod h1:SK9X900oXmPWilYR5/WKPzt3Kqxn/uS/+lbpREv+eCg=
github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/spkg/bom v0.0.0-20160624110644-59b7046e48ad/go.mod h1:qLr4V1qq6nMqFKkMo8ZTx3f+BZEkzsRUY10Xsm2mwU0=
github.com/stretchr/objx v0.1.0/go.mod h1:HFkY916IF+rwdDfMAkV7OtwuqBVzrE8GR6GFx+wExME=
github.com/stretchr/testify v1.3.0/go.mod h1:M5WIy9Dh21IEIfnGCwXGc5bZfKNJtfHm1UVUgZn+9EI=
github.com/stretchr/testify v1.8.4 h1:CcVxjf3Q8PM0mHUKJCdn+eZZtm5yQwehR5yeSVQQcUk=
github.com/stretchr/testify v1.8.4/go.mod h1:sz/lmYIOXD/1dqDmKjjqLyZ2RngseejIcXlSw2iwfAo=
github.com/windmill-labs/windmill-go-client v1.501.4 h1:uwaWmLgZ0JXNusbp/j8r0DQflF62j/SITOhNcBJEYSg=
github.com/windmill-labs/windmill-go-client v1.501.4/go.mod h1:RjnzKr3lc7/Gr72zXDmmqtnUm6PzU+UVHhd3alNKxoM=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=`,
			Type: "rawscript",
			Content: `package inner

import (
	"fmt"
	"net/http"
  "encoding/json"
  "bytes"

	wmill "github.com/windmill-labs/windmill-go-client"
)


func main(x string, nested struct {
	Foo string ` + "`json:\"foo\"`" + `
}) (interface{}, error) {
	v, err := wmill.GetVariable("u/admin/extraordinary_variable")
	if err != nil {
		return nil, err
	}
	fmt.Println("Fetched variable:", v)

	payload := struct {
		Status string ` + "`json:\"status\"`" + `
		Token  string ` + "`json:\"token\"`" + `
	}{
		Status: "success",
		Token:  v,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		"https://play.svix.com/in/e_kZU5EMJbZYBBWXOOK2qsuqi8vl0/",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return x, nil
}`,
			Language: "go",
			InputTransforms: struct {
				X struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"x"`
				Nested struct {
					Type  string `json:"type"`
					Value struct {
						Foo string `json:"foo"`
					} `json:"value"`
				} `json:"nested"`
			}{
				X: struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				}{
					Type:  "static",
					Value: "",
				},
				Nested: struct {
					Type  string `json:"type"`
					Value struct {
						Foo string `json:"foo"`
					} `json:"value"`
				}{
					Type: "static",
					Value: struct {
						Foo string `json:"foo"`
					}{
						Foo: "",
					},
				},
			},
		},
	}

	return append(modules, windmillModule)
}

// generateRandomID generates a random string ID
func generateRandomID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
