package buildkite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/theopenlane/core/internal/integrations/providers/helpers"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	buildkiteOperationHealth types.OperationName = "health.default"
	buildkiteOperationOrgs   types.OperationName = "organizations.collect"
)

var buildkiteHTTPClient = &http.Client{Timeout: 10 * time.Second}

func buildkiteOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        buildkiteOperationHealth,
			Kind:        types.OperationKindHealth,
			Description: "Validate Buildkite token by calling the /v2/user endpoint.",
			Run:         runBuildkiteHealthOperation,
		},
		{
			Name:        buildkiteOperationOrgs,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Buildkite organizations for reporting.",
			Run:         runBuildkiteOrganizationsOperation,
		},
	}
}

type buildkiteUserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type buildkiteOrgResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	WebURL string `json:"web_url"`
}

func runBuildkiteHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.APITokenFromPayload(input.Credential, string(TypeBuildkite))
	if err != nil {
		return types.OperationResult{}, err
	}

	var user buildkiteUserResponse
	if err := buildkiteAPIGet(ctx, token, "user", &user); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Buildkite user lookup failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Buildkite token valid for %s", user.Name),
		Details: map[string]any{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"username": user.Username,
		},
	}, nil
}

func runBuildkiteOrganizationsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.APITokenFromPayload(input.Credential, string(TypeBuildkite))
	if err != nil {
		return types.OperationResult{}, err
	}

	var orgs []buildkiteOrgResponse
	if err := buildkiteAPIGet(ctx, token, "organizations", &orgs); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Buildkite organizations fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	samples := make([]map[string]any, 0, min(5, len(orgs)))
	for _, org := range orgs {
		if len(samples) >= cap(samples) {
			break
		}
		samples = append(samples, map[string]any{
			"id":   org.ID,
			"name": org.Name,
			"slug": org.Slug,
			"url":  org.WebURL,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Discovered %d Buildkite organizations", len(orgs)),
		Details: map[string]any{
			"count":   len(orgs),
			"samples": samples,
		},
	}, nil
}

func buildkiteAPIGet(ctx context.Context, token, path string, out any) error {
	endpoint := fmt.Sprintf("https://api.buildkite.com/v2/%s", path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := buildkiteHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("buildkite api %s: %s", path, resp.Status)
	}

	if out == nil {
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
