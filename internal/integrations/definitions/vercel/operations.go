package vercel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// UserResponse is the raw Vercel /v2/user response shape
type UserResponse struct {
	User struct {
		ID    string `json:"uid"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

// HealthCheck holds the result of a Vercel health check
type HealthCheck struct {
	// ID is the Vercel user identifier
	ID string `json:"id"`
	// Name is the Vercel user display name
	Name string `json:"name"`
	// Email is the Vercel user email
	Email string `json:"email"`
}

// ProjectSample holds a single Vercel project entry
type ProjectSample struct {
	// ID is the Vercel project identifier
	ID string `json:"id"`
	// Name is the Vercel project name
	Name string `json:"name"`
	// Framework is the detected project framework
	Framework string `json:"framework"`
}

// ProjectsSample collects and returns a sample of Vercel projects
type ProjectsSample struct {
	// Projects is the collected project sample
	Projects []ProjectSample `json:"projects"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, c)
	}
}

// Run executes the Vercel health check
func (HealthCheck) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	var resp UserResponse
	if err := c.GetJSON(ctx, "/v2/user", &resp); err != nil {
		return nil, fmt.Errorf("vercel: user lookup failed: %w", err)
	}

	return jsonx.ToRawMessage(HealthCheck{
		ID:    resp.User.ID,
		Name:  resp.User.Name,
		Email: resp.User.Email,
	})
}

// Handle adapts projects sample to the generic operation registration boundary
func (p ProjectsSample) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return p.Run(ctx, c)
	}
}

// Run fetches a small sample of Vercel projects
func (ProjectsSample) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("limit", "5")

	var resp struct {
		Projects []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Framework string `json:"framework"`
		} `json:"projects"`
	}

	if err := c.GetJSONWithParams(ctx, "/v4/projects", params, &resp); err != nil {
		return nil, fmt.Errorf("vercel: projects fetch failed: %w", err)
	}

	samples := make([]ProjectSample, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		samples = append(samples, ProjectSample{
			ID:        p.ID,
			Name:      p.Name,
			Framework: p.Framework,
		})
	}

	return jsonx.ToRawMessage(ProjectsSample{Projects: samples})
}
